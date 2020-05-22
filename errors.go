package chatter

import (
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

var (
    CLIENT_NOT_FOUND error = status.Error(codes.NotFound, "client not found")
    INVALID_NICKNAME error = status.Error(codes.InvalidArgument, "invalid nickname")
    NICKNAME_USED error = status.Error(codes.AlreadyExists, "nickname is already used by another client")
)

