package mppj

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/grpc/metadata"
)

type SourceID string

type contextKey string

const sourceIDContextKey = contextKey("source-id")

func SourceIDToOutgoingContext(ctx context.Context, id SourceID) context.Context {
	return metadata.AppendToOutgoingContext(context.Background(), string(sourceIDContextKey), string(id))
}

func SourceIDFromIncomingContext(ctx context.Context) (SourceID, bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", false
	}
	id := md.Get(string(sourceIDContextKey))
	if len(id) == 0 {
		return "", false
	}
	return SourceID(id[0]), true
}

type SourceList []SourceID

func (s *SourceList) String() string {
	return fmt.Sprintf("%v", *s)
}

func (s *SourceList) Set(value string) error {
	ids := strings.Split(value, ",")
	for _, id := range ids {
		*s = append(*s, SourceID(id))
	}
	return nil
}
