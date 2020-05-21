package main

import (
    "errors"
)

var (
    CLIENT_NOT_FOUND error = errors.New("client not found")
    INVALID_NICKNAME error = errors.New("invalid nickname")
    NICKNAME_USED error = errors.New("nickname is already used by another client")
)

