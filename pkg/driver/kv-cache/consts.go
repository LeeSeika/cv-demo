package kvcache

import "time"

var emptyValuePlaceholder = []byte("__EMPTY VALUE PLACEHOLDER__")

const emptyValuePlaceholderTTL = 2 * time.Minute
