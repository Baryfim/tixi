package grpc

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/tixiby/api/proto/authpb"
	"github.com/tixiby/internal/config"
	"github.com/tixiby/pkg/auth"
)

func RunGRPCServer() error {
	cert, err := tls.LoadX509KeyPair(config.Cfg.SSLCert, config.Cfg.SSLKey)
	if err != nil {
		return err
	}

	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile(config.Cfg.SSLCert)
	if err != nil {
		return err
	}
	certPool.AppendCertsFromPEM(ca)

	creds := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{cert},
	})

	lis, err := net.Listen("tcp", config.Cfg.GRPCPort)
	if err != nil {
		return err
	}

	s := grpc.NewServer(grpc.Creds(creds))
	authpb.RegisterAuthServiceServer(s, &auth.AuthServiceServer{})

	log.Printf("gRPC сервер запущен на порту %s", config.Cfg.GRPCPort)

	if err := s.Serve(lis); err != nil {
		return err
	}

	return nil
}
