package vectordb

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/leon37/FaceTaxLedger/internal/infrastructure/embedding"
	"github.com/leon37/FaceTaxLedger/internal/repository"
	pb "github.com/qdrant/go-client/qdrant"
	"log"
)

type QdrantRepository struct {
	client   *QdrantClient // 持有我们在上一节写的底层 client
	embedder embedding.Provider
}

func (r *QdrantRepository) SaveMemory(ctx context.Context, uid string, expenseID uint, description string) error {
	// 1. 将文本转化为向量 (调用 Embedding API)
	vec, err := r.embedder.GetVector(ctx, description)
	if err != nil {
		return fmt.Errorf("failed to vectorize text: %v", err)
	}

	// 2. 构造 Qdrant Point
	// Point ID 这里我们临时生成一个 UUID，或者你可以复用 expenseID (需要转 uint64)
	// 为了简单，我们这里用 UUID 字符串作为 Point ID
	pointID := uuid.New().String()

	points := []*pb.PointStruct{
		{
			Id: &pb.PointId{
				PointIdOptions: &pb.PointId_Uuid{Uuid: pointID},
			},
			Vectors: &pb.Vectors{
				VectorsOptions: &pb.Vectors_Vector{
					Vector: &pb.Vector{Data: vec},
				},
			},
			Payload: map[string]*pb.Value{
				"user_id":     {Kind: &pb.Value_StringValue{StringValue: uid}},
				"expense_id":  {Kind: &pb.Value_IntegerValue{IntegerValue: int64(expenseID)}},
				"description": {Kind: &pb.Value_StringValue{StringValue: description}},
			},
		},
	}

	wait := true
	// 3. Upsert 入库
	_, err = r.client.points.Upsert(ctx, &pb.UpsertPoints{
		CollectionName: CollectionName,
		Points:         points,
		Wait:           &wait,
	})

	if err != nil {
		return fmt.Errorf("qdrant upsert failed: %v", err)
	}

	log.Printf("Saved memory to Qdrant. ID: %d, Desc: %s", expenseID, description)
	return nil
}

func (r *QdrantRepository) SearchSimilar(ctx context.Context, uid string, description string, limit int) ([]string, error) {
	// 1. 将查询文本转化为向量
	queryVector, err := r.embedder.GetVector(ctx, description)
	if err != nil {
		return nil, fmt.Errorf("embedding failed: %v", err)
	}

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
		return nil, fmt.Errorf("qdrant search failed: %v", err)
	}

	// 3. 提取结果中的文本
	var histories []string
	for _, point := range searchResult.Result {
		// point.Score 是相似度分数，你可以选择过滤掉分数太低的（比如 < 0.7）
		// 这里暂不设限，全部通过

		// 从 Payload Map 中获取 description 字段
		if val, ok := point.Payload["description"]; ok {
			// 使用 Type Assertion 提取 StringValue
			// Protobuf 的 Value 结构很深: Value -> Kind(oneof) -> StringValue
			if strVal, ok := val.Kind.(*pb.Value_StringValue); ok {
				histories = append(histories, strVal.StringValue)
			}
		}
	}

	return histories, nil
}

// NewQdrantRepository 构造函数
func NewQdrantRepository(client *QdrantClient, embedder embedding.Provider) repository.MemoryRepo {
	return &QdrantRepository{
		client:   client,
		embedder: embedder,
	}
}
