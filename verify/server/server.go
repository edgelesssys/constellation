package server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"

	"github.com/edgelesssys/constellation/internal/logger"
	"github.com/edgelesssys/constellation/verify/verifyproto"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type attestation struct {
	Data []byte `json:"data"`
}

// Server implements Constellation's verify API.
// The server exposes both HTTP and gRPC endpoints
// to retrieve attestation statements.
type Server struct {
	log    *logger.Logger
	issuer AttestationIssuer
	verifyproto.UnimplementedAPIServer
}

// New initializes a new verification server.
func New(log *logger.Logger, issuer AttestationIssuer) *Server {
	return &Server{
		log:    log,
		issuer: issuer,
	}
}

// Run starts the HTTP and gRPC servers.
// If one of the servers fails, other server will be closed and the error will be returned.
func (s *Server) Run(httpListener, grpcListener net.Listener) error {
	var err error
	var wg sync.WaitGroup
	var once sync.Once

	s.log.WithIncreasedLevel(zapcore.WarnLevel).Named("grpc").ReplaceGRPCLogger()
	grpcServer := grpc.NewServer(s.log.Named("gRPC").GetServerUnaryInterceptor())
	verifyproto.RegisterAPIServer(grpcServer, s)

	httpHandler := http.NewServeMux()
	httpHandler.HandleFunc("/", s.getAttestationHTTP)
	httpServer := &http.Server{Handler: httpHandler}

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer grpcServer.GracefulStop()

		s.log.Infof("Starting HTTP server on %s", httpListener.Addr().String())
		httpErr := httpServer.Serve(httpListener)
		if httpErr != nil && httpErr != http.ErrServerClosed {
			once.Do(func() { err = httpErr })
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() { _ = httpServer.Shutdown(context.Background()) }()

		s.log.Infof("Starting gRPC server on %s", grpcListener.Addr().String())
		grpcErr := grpcServer.Serve(grpcListener)
		if grpcErr != nil {
			once.Do(func() { err = grpcErr })
		}
	}()

	wg.Wait()
	return err
}

// GetAttestation implements the gRPC endpoint for requesting attestation statements.
func (s *Server) GetAttestation(ctx context.Context, req *verifyproto.GetAttestationRequest) (*verifyproto.GetAttestationResponse, error) {
	peerAddr := "unknown"
	if peer, ok := peer.FromContext(ctx); ok {
		peerAddr = peer.Addr.String()
	}

	log := s.log.With(zap.String("peerAddress", peerAddr)).Named("gRPC")
	s.log.Infof("Received attestation request")
	if len(req.Nonce) == 0 {
		log.Errorf("Received attestation request with empty nonce")
		return nil, status.Error(codes.InvalidArgument, "nonce is required to issue attestation")
	}
	if len(req.UserData) == 0 {
		log.Errorf("Received attestation request with empty user data")
		return nil, status.Error(codes.InvalidArgument, "user data is required to issue attestation")
	}

	log.Infof("Creating attestation")
	statement, err := s.issuer.Issue(req.UserData, req.Nonce)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "issuing attestation statement: %v", err)
	}

	log.Infof("Attestation request successful")
	return &verifyproto.GetAttestationResponse{Attestation: statement}, nil
}

// getAttestationHTTP implements the HTTP endpoint for retrieving attestation statements.
func (s *Server) getAttestationHTTP(w http.ResponseWriter, r *http.Request) {
	log := s.log.With(zap.String("peerAddress", r.RemoteAddr)).Named("http")

	nonceB64 := r.URL.Query()["nonce"]
	if len(nonceB64) != 1 || nonceB64[0] == "" {
		log.Errorf("Received attestation request with empty or multiple nonce parameter")
		http.Error(w, "nonce parameter is required exactly once", http.StatusBadRequest)
		return
	}
	userDataB64 := r.URL.Query()["userData"]
	if len(userDataB64) != 1 || userDataB64[0] == "" {
		log.Errorf("Received attestation request with empty or multiple user data parameter")
		http.Error(w, "userData parameter is required exactly once", http.StatusBadRequest)
		return
	}

	nonce, err := base64.URLEncoding.DecodeString(nonceB64[0])
	if err != nil {
		log.With(zap.Error(err)).Errorf("Received attestation request with invalid nonce")
		http.Error(w, fmt.Sprintf("invalid base64 encoding for nonce: %v", err), http.StatusBadRequest)
		return
	}
	userData, err := base64.URLEncoding.DecodeString(userDataB64[0])
	if err != nil {
		log.With(zap.Error(err)).Errorf("Received attestation request with invalid user data")
		http.Error(w, fmt.Sprintf("invalid base64 encoding for userData: %v", err), http.StatusBadRequest)
		return
	}

	log.Infof("Creating attestation")
	quote, err := s.issuer.Issue(userData, nonce)
	if err != nil {
		http.Error(w, fmt.Sprintf("issuing attestation statement: %v", err), http.StatusInternalServerError)
		return
	}

	log.Infof("Attestation request successful")
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(attestation{quote}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type AttestationIssuer interface {
	Issue(userData []byte, nonce []byte) (quote []byte, err error)
}
