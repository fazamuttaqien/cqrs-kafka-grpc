// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v3.21.12
// source: product_reader.proto

package readerService

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

var File_product_reader_proto protoreflect.FileDescriptor

const file_product_reader_proto_rawDesc = "" +
	"\n" +
	"\x14product_reader.proto\x12\rreaderService\x1a\x1dproduct_reader_messages.proto2\xb4\x03\n" +
	"\rreaderService\x12R\n" +
	"\rCreateProduct\x12\x1f.readerService.CreateProductReq\x1a .readerService.CreateProductResp\x12R\n" +
	"\rUpdateProduct\x12\x1f.readerService.UpdateProductReq\x1a .readerService.UpdateProductResp\x12U\n" +
	"\x0eGetProductById\x12 .readerService.GetProductByIdReq\x1a!.readerService.GetProductByIdResp\x12D\n" +
	"\rSearchProduct\x12\x18.readerService.SearchReq\x1a\x19.readerService.SearchResp\x12^\n" +
	"\x11DeleteProductByID\x12#.readerService.DeleteProductByIdReq\x1a$.readerService.DeleteProductByIdRespB\x12Z\x10./;readerServiceb\x06proto3"

var file_product_reader_proto_goTypes = []any{
	(*CreateProductReq)(nil),      // 0: readerService.CreateProductReq
	(*UpdateProductReq)(nil),      // 1: readerService.UpdateProductReq
	(*GetProductByIdReq)(nil),     // 2: readerService.GetProductByIdReq
	(*SearchReq)(nil),             // 3: readerService.SearchReq
	(*DeleteProductByIdReq)(nil),  // 4: readerService.DeleteProductByIdReq
	(*CreateProductResp)(nil),     // 5: readerService.CreateProductResp
	(*UpdateProductResp)(nil),     // 6: readerService.UpdateProductResp
	(*GetProductByIdResp)(nil),    // 7: readerService.GetProductByIdResp
	(*SearchResp)(nil),            // 8: readerService.SearchResp
	(*DeleteProductByIdResp)(nil), // 9: readerService.DeleteProductByIdResp
}
var file_product_reader_proto_depIdxs = []int32{
	0, // 0: readerService.readerService.CreateProduct:input_type -> readerService.CreateProductReq
	1, // 1: readerService.readerService.UpdateProduct:input_type -> readerService.UpdateProductReq
	2, // 2: readerService.readerService.GetProductById:input_type -> readerService.GetProductByIdReq
	3, // 3: readerService.readerService.SearchProduct:input_type -> readerService.SearchReq
	4, // 4: readerService.readerService.DeleteProductByID:input_type -> readerService.DeleteProductByIdReq
	5, // 5: readerService.readerService.CreateProduct:output_type -> readerService.CreateProductResp
	6, // 6: readerService.readerService.UpdateProduct:output_type -> readerService.UpdateProductResp
	7, // 7: readerService.readerService.GetProductById:output_type -> readerService.GetProductByIdResp
	8, // 8: readerService.readerService.SearchProduct:output_type -> readerService.SearchResp
	9, // 9: readerService.readerService.DeleteProductByID:output_type -> readerService.DeleteProductByIdResp
	5, // [5:10] is the sub-list for method output_type
	0, // [0:5] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_product_reader_proto_init() }
func file_product_reader_proto_init() {
	if File_product_reader_proto != nil {
		return
	}
	file_product_reader_messages_proto_init()
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_product_reader_proto_rawDesc), len(file_product_reader_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_product_reader_proto_goTypes,
		DependencyIndexes: file_product_reader_proto_depIdxs,
	}.Build()
	File_product_reader_proto = out.File
	file_product_reader_proto_goTypes = nil
	file_product_reader_proto_depIdxs = nil
}
