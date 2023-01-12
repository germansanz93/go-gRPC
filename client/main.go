package main

import (
	"context"
	"germansanz93/go/grpc/testpb"
	"io"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	cc, err := grpc.Dial("localhost:5070", grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}

	defer cc.Close()

	c := testpb.NewTestServiceClient(cc)

	// DoUnary(c)
	// DoClientStreaming(c)
	// DoServerStreaming(c)
	DoBiDirectionStreaming(c)
}

func DoUnary(c testpb.TestServiceClient) {
	req := &testpb.GetTestRequest{
		Id: "t1",
	}

	res, err := c.GetTest(context.Background(), req)

	if err != nil {
		log.Fatalf("error while calling GetTest: %v", err)
	}

	log.Printf("response from server: %v", res)
}

func DoClientStreaming(c testpb.TestServiceClient) {
	questions := []*testpb.Question{
		{
			Id:       "q8t1",
			Answer:   "Azul",
			Question: "Color asociado a Go",
			TestId:   "t1",
		},
		{
			Id:       "q9t1",
			Answer:   "Google",
			Question: "Empresa que desarrollo Go",
			TestId:   "t1",
		},
		{
			Id:       "q10t1",
			Answer:   "Backend",
			Question: "Especialidad de Go",
			TestId:   "t1",
		},
	}

	stream, err := c.SetQuestions(context.Background())
	if err != nil {
		log.Fatalf("error while calling SetQuestions: %v", err)
	}

	for _, question := range questions {
		log.Println("sending question: ", question.Id)
		stream.Send(question)
		time.Sleep(2 * time.Second)
	}

	msg, err := stream.CloseAndRecv()

	if err != nil {
		log.Fatalf("error while receiving response: %v", err)
	}
	log.Printf("response from server: %v", msg)

}

func DoServerStreaming(c testpb.TestServiceClient) {
	req := &testpb.GetStudentsPerTestRequest{
		TestId: "t1",
	}

	stream, err := c.GetStudentsPerTest(context.Background(), req)

	if err != nil {
		log.Fatalf("error while calling GetStudentsPerTest: %v", err)
	}

	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("error while reading stream: %v", err)
		}
		log.Printf("response from server: %v", msg)
	}
}

func DoBiDirectionStreaming(c testpb.TestServiceClient) {
	answer := testpb.TakeTestRequest{
		Answer: "42",
	}

	numberOfQuestions := 4

	waitChannel := make(chan struct{})

	stream, err := c.TakeTest(context.Background())
	if err != nil {
		log.Fatalf("error while calling TakeTest: %v", err)
	}

	go func() {
		for i := 0; i < numberOfQuestions; i++ {
			stream.Send(&answer)
			time.Sleep(2 * time.Second)
		}
	}()

	go func() {
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("error while reading stream: %v", err)
				break
			}
			log.Printf("response from server: %v", res)
		}
		close(waitChannel)
	}()
	<-waitChannel
}
