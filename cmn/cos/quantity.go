// Package cos provides common low-level types and utilities for all aistore projects
/*
 * Copyright (c) 2018-2021, NVIDIA CORPORATION. All rights reserved.
 */
package cos

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

const (
	QuantityPercent = "percent"
	QuantityBytes   = "bytes"
)

type (
	ParsedQuantity struct {
		Type  string
		Value uint64
	}
)

///////////////////
// ParseQuantity //
///////////////////

func ParseQuantity(quantity string) (ParsedQuantity, error) {
	quantity = strings.ReplaceAll(quantity, " ", "")
	idx := 0
	number := ""
	for ; idx < len(quantity) && unicode.IsDigit(rune(quantity[idx])); idx++ {
		number += string(quantity[idx])
	}

	parsedQ := ParsedQuantity{}
	if value, err := strconv.Atoi(number); err != nil {
		return parsedQ, ErrQuantityUsage
	} else if value < 0 {
		return parsedQ, errQuantityNonNegative
	} else {
		parsedQ.Value = uint64(value)
	}

	if len(quantity) <= idx {
		return parsedQ, ErrQuantityUsage
	}

	suffix := quantity[idx:]
	if suffix == "%" {
		parsedQ.Type = QuantityPercent
		if parsedQ.Value == 0 || parsedQ.Value >= 100 {
			return parsedQ, ErrQuantityPercent
		}
	} else if value, err := ParseSize(quantity, UnitsIEC); err != nil {
		return parsedQ, err
	} else if value < 0 {
		return parsedQ, ErrQuantityBytes
	} else {
		parsedQ.Type = QuantityBytes
		parsedQ.Value = uint64(value)
	}

	return parsedQ, nil
}

func (pq ParsedQuantity) String() string {
	switch pq.Type {
	case QuantityPercent:
		return fmt.Sprintf("%d%%", pq.Value)
	case QuantityBytes:
		return ToSizeIEC(int64(pq.Value), 2)
	default:
		AssertMsg(false, fmt.Sprintf("Unknown quantity type: %s", pq.Type))
		return ""
	}
}
