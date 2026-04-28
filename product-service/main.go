package main

import (
	"context"
	"log"
	"net"

	pb "github.com/harun1804/micro-grpc-go/proto/pb" // sesuaikan path module Anda
	"google.golang.org/grpc"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Product struct {
	ID    string `gorm:"primaryKey"`
	Name  string
	Price float32
}

type server struct {
	pb.UnimplementedProductServiceServer
	db *gorm.DB
}

func (s *server) GetProducts(in *pb.Empty, stream pb.ProductService_GetProductsServer) error {
	var products []Product
	if err := s.db.Find(&products).Error; err != nil {
		return err
	}

	for _, p := range products {
		if err := stream.Send(&pb.ProductResponse{
			Id:    p.ID,
			Name:  p.Name,
			Price: p.Price,
		}); err != nil {
			return err
		}
	}

	return nil
}

func (s *server) GetProduct(ctx context.Context, in *pb.ProductRequest) (*pb.ProductResponse, error) {
	var product Product
	if err := s.db.First(&product, "id = ?", in.Id).Error; err != nil {
		return nil, err
	}

	return &pb.ProductResponse{
		Id:    product.ID,
		Name:  product.Name,
		Price: product.Price,
	}, nil
}

// Create
func (s *server) CreateProduct(ctx context.Context, in *pb.ProductCreateRequest) (*pb.ProductResponse, error) {
	p := Product{ID: in.Id, Name: in.Name, Price: in.Price}
	if err := s.db.Create(&p).Error; err != nil {
		return nil, err
	}
	return &pb.ProductResponse{Id: p.ID, Name: p.Name, Price: p.Price}, nil
}

// Update
func (s *server) UpdateProduct(ctx context.Context, in *pb.ProductUpdateRequest) (*pb.ProductResponse, error) {
	var p Product
	if err := s.db.First(&p, "id = ?", in.Id).Error; err != nil {
		return nil, err
	}
	p.Name = in.Name
	p.Price = in.Price
	s.db.Save(&p)
	return &pb.ProductResponse{Id: p.ID, Name: p.Name, Price: p.Price}, nil
}

// Delete
func (s *server) DeleteProduct(ctx context.Context, in *pb.ProductRequest) (*pb.DeleteResponse, error) {
	if err := s.db.Delete(&Product{}, "id = ?", in.Id).Error; err != nil {
		return nil, err
	}
	return &pb.DeleteResponse{Message: "Produk berhasil dihapus"}, nil
}

func main() {
	// Koneksi MySQL
	dsn := "root:root@tcp(127.0.0.1:3307)/microservice_db?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Gagal koneksi DB: %v", err)
	}

	// Auto-migrate tabel
	db.AutoMigrate(&Product{})

	// Setup gRPC Server
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Gagal listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterProductServiceServer(s, &server{db: db})

	log.Println("Product Service (gRPC + MySQL) berjalan di port :50051...")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Gagal serve: %v", err)
	}
}
