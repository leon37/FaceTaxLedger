package vectordb

import (
	"context"
	"fmt"
	pb "github.com/qdrant/go-client/qdrant"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
)

const (
	CollectionName = "facetax_memory"
	VectorSize     = 1536 // 待定：根据后续选用的 Embedding 模型维度修改
)

type QdrantClient struct {
	conn   *grpc.ClientConn
	client pb.CollectionsClient
	points pb.PointsClient
}

// NewQdrantClient 初始化连接
func NewQdrantClient(host string, port int) (*QdrantClient, error) {
	addr := fmt.Sprintf("%s:%d", host, port)

	// 连接 Qdrant (gRPC)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("did not connect to qdrant: %v", err)
	}

	return &QdrantClient{
		conn:   conn,
		client: pb.NewCollectionsClient(conn),
		points: pb.NewPointsClient(conn),
	}, nil
}

// Close 关闭连接
func (q *QdrantClient) Close() {
	if q.conn != nil {
		q.conn.Close()
	}
}

// InitCollection 确保向量集合存在
func (q *QdrantClient) InitCollection(ctx context.Context) error {
	// 1. 检查集合是否存在
	exists, err := q.client.Get(ctx, &pb.GetCollectionInfoRequest{
		CollectionName: CollectionName,
	})

	// 如果不出错且存在，则直接返回
	if err == nil && exists != nil {
		log.Printf("Qdrant Collection '%s' already exists.", CollectionName)
		return nil
	}

	// 2. 如果不存在，创建集合
	log.Printf("Creating Qdrant Collection '%s' with dim %d...", CollectionName, VectorSize)

	_, err = q.client.Create(ctx, &pb.CreateCollection{
		CollectionName: CollectionName,
		VectorsConfig: &pb.VectorsConfig{
			Config: &pb.VectorsConfig_Params{
				Params: &pb.VectorParams{
					Size:     VectorSize,
					Distance: pb.Distance_Cosine, // 余弦相似度最适合文本语义检索
				},
			},
		},
	})

	if err != nil {
		return fmt.Errorf("failed to create collection: %v", err)
	}

	log.Println("Qdrant Collection created successfully.")
	return nil
}
