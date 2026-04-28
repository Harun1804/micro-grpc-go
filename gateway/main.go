package main

import (
	"context"
	"log"
	"time"

	"github.com/gofiber/fiber/v3"
	pb "github.com/harun1804/micro-grpc-go/proto/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	app := fiber.New()

	// Setup koneksi gRPC ke Product Service
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Tidak bisa konek ke gRPC service: %v", err)
	}
	defer conn.Close()
	client := pb.NewProductServiceClient(conn)

	app.Get("/product", func(c fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		res, err := client.GetProducts(ctx, &pb.Empty{})
		if err != nil {
			return c.Status(500).SendString("Gagal mengambil produk: " + err.Error())
		}

		var products []fiber.Map
		for {
			p, err := res.Recv()
			if err != nil {
				break
			}
			products = append(products, fiber.Map{
				"id":    p.Id,
				"name":  p.Name,
				"price": p.Price,
			})
		}

		return c.JSON(products)
	})

	// Endpoint Fiber
	app.Get("/product/:id", func(c fiber.Ctx) error {
		id := c.Params("id")

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		// Panggil Service via gRPC
		res, err := client.GetProduct(ctx, &pb.ProductRequest{Id: id})
		if err != nil {
			return c.Status(404).JSON(fiber.Map{
				"message": "Produk tidak ditemukan atau service mati",
			})
		}

		return c.JSON(fiber.Map{
			"id":    res.Id,
			"name":  res.Name,
			"price": res.Price,
		})
	})

	// Create Product
	app.Post("/product", func(c fiber.Ctx) error {
		req := new(pb.ProductCreateRequest)
		if err := c.Bind().Body(req); err != nil {
			return err
		}
		res, err := client.CreateProduct(context.Background(), req)
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}
		return c.JSON(res)
	})

	// Update Product
	app.Put("/product/:id", func(c fiber.Ctx) error {
		req := new(pb.ProductUpdateRequest)
		if err := c.Bind().Body(req); err != nil {
			return err
		}
		req.Id = c.Params("id")
		res, err := client.UpdateProduct(context.Background(), req)
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}
		return c.JSON(res)
	})

	// Delete Product
	app.Delete("/product/:id", func(c fiber.Ctx) error {
		res, err := client.DeleteProduct(context.Background(), &pb.ProductRequest{
			Id: c.Params("id"),
		})
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}
		return c.JSON(res)
	})

	log.Fatal(app.Listen(":3000"))
}
