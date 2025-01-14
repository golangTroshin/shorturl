package grpc

import (
	"context"
	"log"

	"github.com/golangTroshin/shorturl/internal/app/helpers"
	"github.com/golangTroshin/shorturl/internal/app/http/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// GiveAuthTokenToUserInterceptor is a gRPC interceptor that assigns an authentication token to the user.
//
// If the token is missing in the metadata, it generates a new token and adds it to the context and metadata.
//
// Parameters:
//   - ctx: The context for the request.
//   - req: The gRPC request.
//   - info: Details about the gRPC method being called.
//   - handler: The next handler in the interceptor chain.
//
// Returns:
//   - The response from the next handler or an error if token generation fails.
func GiveAuthTokenToUserInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	}

	var authToken string
	if tokens := md[middleware.CookieAuthToken]; len(tokens) > 0 {
		authToken = tokens[0]
		log.Println("Token found in metadata")
	} else {
		var err error
		authToken, err = helpers.BuildJWTString()
		if err != nil {
			log.Printf("BuildJWTString error: %v", err)
			return nil, status.Errorf(codes.Internal, "Internal Server Error")
		}
		log.Println("Generated new auth token")
	}

	// Add token to the context
	ctx = context.WithValue(ctx, middleware.UserIDKey, authToken)
	outgoingMD := metadata.Pairs(middleware.CookieAuthToken, authToken)
	grpc.SetHeader(ctx, outgoingMD)

	if authToken != "" {
		// Inject the token into the outgoing metadata
		md, _ := metadata.FromIncomingContext(ctx)
		md = metadata.Join(md, outgoingMD)
		ctx = metadata.NewIncomingContext(ctx, md)
	}

	// Call the next handler
	return handler(ctx, req)
}

// CheckAuthTokenInterceptor validates the presence and correctness of the auth token.
//
// Parameters:
//   - ctx: The context for the request.
//   - req: The gRPC request.
//   - info: Details about the gRPC method being called.
//   - handler: The next handler in the interceptor chain.
//
// Returns:
//   - The response from the next handler or an error if token validation fails.
func CheckAuthTokenInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "No metadata found")
	}

	// Check for auth_token in metadata
	tokens := md[middleware.CookieAuthToken]
	if len(tokens) == 0 || tokens[0] == "" {
		log.Println("Missing or empty auth_token")
		return nil, status.Errorf(codes.Unauthenticated, "Invalid or missing auth token")
	}

	authToken := tokens[0]
	log.Printf("Valid auth token: %s", authToken)

	// Add the token to the context for downstream handlers
	ctx = context.WithValue(ctx, middleware.UserIDKey, authToken)

	// Call the next handler
	return handler(ctx, req)
}
