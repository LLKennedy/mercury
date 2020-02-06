package proxy

import (
	"context"
	"fmt"
	"reflect"

	"google.golang.org/grpc"
)

// Beware, this is where stuff gets super vague thanks to the magic of reflection

func validateArgs(expected, found reflect.Type) error {
	// All this to get a proper array out of these reflection types
	expectedInLen := expected.NumIn()
	expectedOutLen := expected.NumOut()
	foundInLen := found.NumIn()
	foundOutLen := found.NumOut()
	expectedIn := []reflect.Type{}
	for i := 0; i < expectedInLen; i++ {
		expectedIn = append(expectedIn, expected.In(i))
	}
	expectedOut := []reflect.Type{}
	for i := 0; i < expectedOutLen; i++ {
		expectedOut = append(expectedOut, expected.Out(i))
	}
	foundIn := []reflect.Type{}
	for i := 0; i < foundInLen; i++ {
		foundIn = append(foundIn, found.In(i))
	}
	foundOut := []reflect.Type{}
	for i := 0; i < foundOutLen; i++ {
		foundOut = append(foundOut, found.Out(i))
	}
	if expectedInLen < 2 || foundInLen < 2 {
		return fmt.Errorf("cannot exclude receiver from argument checks if receiver is the only argument: expected >= 2 input argments, found %d and %d", expectedInLen, foundInLen)
	}
	if !isStructPtr(expectedIn[0]) || !isStructPtr(foundIn[0]) {
		return fmt.Errorf("no receiver")
	}
	// Don't check receivers, those don't have to be the same type
	err := typesMatch(expectedIn[1:], foundIn[1:])
	if err != nil {
		return err
	}
	err = typesMatch(expectedOut, foundOut)
	return err
}

func typesMatch(expected, found []reflect.Type) error {
	if len(expected) != len(found) {
		return fmt.Errorf("argument lengths did not match: expected %d but found %d", len(expected), len(found))
	}
	for i := range expected {
		if expected[i].Kind() != found[i].Kind() {
			return fmt.Errorf("argments mismatch in position %d: %s vs %s", i, expected[i].Kind(), found[i].Kind())
		}
	}
	return nil
}

// isStructPtr returns true if the pointer stack exists and resolves to a struct
func isStructPtr(in reflect.Type) bool {
	for in.Kind() == reflect.Ptr {
		in = in.Elem()
		if in.Kind() == reflect.Struct {
			return true
		}
	}
	return false
}

func isContext(in reflect.Type) bool {
	return in.Implements(reflect.TypeOf((*context.Context)(nil)).Elem())
}

func isError(in reflect.Type) bool {
	return in.Implements(reflect.TypeOf((*error)(nil)).Elem())
}

func isOutStream(in reflect.Type) bool {
	sendMethod, exists := in.MethodByName("Send")
	if !exists {
		return false
	}
	send := sendMethod.Type
	return in.Implements(reflect.TypeOf((*grpc.ServerStream)(nil)).Elem()) && send.NumIn() == 1 && send.NumOut() == 1 && isStructPtr(send.In(0)) && isError(send.Out(0))
}

func isInStream(in reflect.Type) bool {
	recvMethod, exists := in.MethodByName("Recv")
	if !exists {
		return false
	}
	recv := recvMethod.Type
	return in.Implements(reflect.TypeOf((*grpc.ServerStream)(nil)).Elem()) && recv.NumIn() == 0 && recv.NumOut() == 2 && isStructPtr(recv.Out(0)) && isError(recv.Out(1))
}

// SendAndClose only applies to StreamStruct patterns
func hasSendAndClose(in reflect.Type) bool {
	sendCloseMethod, exists := in.MethodByName("SendAndClose")
	if !exists {
		return false
	}
	send := sendCloseMethod.Type
	return send.NumIn() == 1 && send.NumOut() == 1 && isStructPtr(send.In(0)) && isError(send.Out(0))
}

func getPattern(args reflect.Type) (pattern apiMethodPattern) {
	defer func() {
		if r := recover(); r != nil {
			// Panic means something wasn't expected, which means this isn't a known pattern
			pattern = apiMethodPatternUnknown
		}
	}()
	// The defer above means we can freely access arguments without checking lengths, as long as it complies with all patterns
	if isStructPtr(args.In(0)) {
		// Pointer receiver checked, filter by first input argument type
		switch {
		case isContext(args.In(1)):
			// We've got an explicit context, this can only be StructStruct, now we just need to confirm
			if args.NumIn() == 3 && isStructPtr(args.In(2)) && args.NumOut() == 2 && isStructPtr(args.Out(0)) && isError(args.Out(1)) {
				pattern = apiMethodPatternStructStruct
			}
		case isStructPtr(args.In(1)):
			// This can only be StructStream
			if args.NumIn() == 3 && isOutStream(args.In(2)) && args.NumOut() == 1 && isError(args.Out(0)) {
				pattern = apiMethodPatternStructStream
			}
		case isInStream(args.In(1)):
			// Either StreamStruct or StreamStream
			if args.NumIn() == 2 && args.NumOut() == 1 && isError(args.Out(0)) {
				switch {
				case hasSendAndClose(args.In(1)):
					// StreamStruct
					pattern = apiMethodPatternStreamStruct
				case isOutStream(args.In(1)):
					// StreamStream
					pattern = apiMethodPatternStreamStream
				}
			}
		}
	}
	return
}
