//go:build darwin

package main

/*
#cgo darwin CFLAGS: -x objective-c
#cgo darwin LDFLAGS: -framework AppKit -framework Foundation

#import <AppKit/AppKit.h>

void copyImageToClipboardDarwin(const unsigned char *data, int len) {
	@autoreleasepool {
		NSData *imgData = [NSData dataWithBytesNoCopy:(void *)data length:len freeWhenDone:NO];
		NSImage *image = [[NSImage alloc] initWithData:imgData];
		if (image) {
			NSPasteboard *pb = [NSPasteboard generalPasteboard];
			[pb clearContents];
			[pb writeObjects:@[image]];
			[image release];
		}
	}
}
*/
import "C"
import "unsafe"

func copyImageToClipboard(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	C.copyImageToClipboardDarwin((*C.uchar)(unsafe.Pointer(&data[0])), C.int(len(data)))
	return nil
}
