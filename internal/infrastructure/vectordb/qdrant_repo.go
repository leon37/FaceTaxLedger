package vectordb

import (
	"context"
	"fmt"
	"github.com/leon37/FaceTaxLedger/internal/repository"
	pb "github.com/qdrant/go-client/qdrant"
	"log"
	"log/slog"
	"time"
)

type QdrantRepository struct {
	client *QdrantClient // 持有我们在上一节写的底层 client

}

func (r *QdrantRepository) SaveMemory(ctx context.Context, uid string, expenseID uint, description string, vector []float32) error {

	// 2. 构造 Qdrant Point

	points := []*pb.PointStruct{
		{
			Id: &pb.PointId{
				PointIdOptions: &pb.PointId_Num{Num: uint64(expenseID)},
			},
			Vectors: &pb.Vectors{
				VectorsOptions: &pb.Vectors_Vector{
					Vector: &pb.Vector{Data: vector},
				},
			},
			Payload: map[string]*pb.Value{
				"user_id":     {Kind: &pb.Value_StringValue{StringValue: uid}},
				"expense_id":  {Kind: &pb.Value_IntegerValue{IntegerValue: int64(expenseID)}},
				"description": {Kind: &pb.Value_StringValue{StringValue: description}},
				"timestamp":   {Kind: &pb.Value_IntegerValue{IntegerValue: time.Now().Unix()}},
			},
		},
	}

	wait := true
	// 3. Upsert 入库
	_, err := r.client.points.Upsert(ctx, &pb.UpsertPoints{
		CollectionName: CollectionName,
		Points:         points,
		Wait:           &wait,
	})

	if err != nil {
		slog.Error("qdrant upsert failed: %v", err)
		return fmt.Errorf("qdrant upsert failed: %v", err)
	}

	log.Printf("Saved memory to Qdrant. ID: %d, Desc: %s", expenseID, description)
	return nil
}

func (r *QdrantRepository) SearchSimilar(ctx context.Context, uid string, limit int, queryVector []float32) ([]repository.MemoryResult, error) {
	condition := &pb.Condition_Field{
		Field: &pb.FieldCondition{
			Key: "user_id", // Qdrant 中的字段名
			Match: &pb.Match{
				MatchValue: &pb.Match_Text{Text: uid},
			},
		},
	}
	filter := &pb.Filter{
		Must: []*pb.Condition{{
			ConditionOneOf: condition,
		}},
	}

	// 2. 调用 Qdrant Search 接口
	// 注意：这里用 limit + 1 是为了预留冗余，或者你可以根据 Score 阈值过滤
	searchResult, err := r.client.points.Search(ctx, &pb.SearchPoints{
		CollectionName: CollectionName,
		Vector:         queryVector,
		Limit:          uint64(limit),
		Filter:         filter,
		// 【关键】必须开启 Enable，否则只返回 ID 和 Score，不返回文本内容
		WithPayload: &pb.WithPayloadSelector{
			SelectorOptions: &pb.WithPayloadSelector_Enable{
				Enable: true,
			},
		},
	})
	if err != nil {
		slog.Error("qdrant search failed: %v", err)
		return nil, fmt.Errorf("qdrant search failed: %v", err)
	}

	// 3. 提取结果中的文本
	var histories []repository.MemoryResult
	for _, point := range searchResult.Result {
		var content string
		var ts int64
		// point.Score 是相似度分数，你可以选择过滤掉分数太低的（比如 < 0.7）
		// 这里暂不设限，全部通过

		// 从 Payload Map 中获取 description 字段
		if val, ok := point.Payload["description"]; ok {
			// 使用 Type Assertion 提取 StringValue
			// Protobuf 的 Value 结构很深: Value -> Kind(oneof) -> StringValue
			if strVal, ok := val.Kind.(*pb.Value_StringValue); ok {
				content = strVal.StringValue
			}
		}
		if t, ok := point.Payload["timestamp"]; ok {
			ts = t.GetIntegerValue()
		}
		histories = append(histories, repository.MemoryResult{
			Content:   content,
			Timestamp: ts,
		})

	}

	return histories, nil
}

func (r *QdrantRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.client.points.Delete(ctx, &pb.DeletePoints{
		CollectionName: CollectionName,
		Points: &pb.PointsSelector{
			PointsSelectorOneOf: &pb.PointsSelector_Points{
				Points: &pb.PointsIdsList{
					Ids: []*pb.PointId{
						{PointIdOptions: &pb.PointId_Num{Num: uint64(id)}}, // 指定要删除的 ID
					},
				},
			},
		},
	})
	return err
}

// NewQdrantRepository 构造函数
func NewQdrantRepository(client *QdrantClient) repository.MemoryRepo {
	return &QdrantRepository{
		client: client,
	}
}
