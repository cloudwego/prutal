// Code generated from ./Protobuf.g4 by ANTLR 4.13.2. DO NOT EDIT.

package parser // Protobuf

import "github.com/cloudwego/prutal/prutalgen/internal/antlr"

// ProtobufListener is a complete listener for a parse tree produced by ProtobufParser.
type ProtobufListener interface {
	antlr.ParseTreeListener

	// EnterProto is called when entering the proto production.
	EnterProto(c *ProtoContext)

	// EnterEdition is called when entering the edition production.
	EnterEdition(c *EditionContext)

	// EnterImportStatement is called when entering the importStatement production.
	EnterImportStatement(c *ImportStatementContext)

	// EnterPackageStatement is called when entering the packageStatement production.
	EnterPackageStatement(c *PackageStatementContext)

	// EnterOptionStatement is called when entering the optionStatement production.
	EnterOptionStatement(c *OptionStatementContext)

	// EnterOptionName is called when entering the optionName production.
	EnterOptionName(c *OptionNameContext)

	// EnterFieldLabel is called when entering the fieldLabel production.
	EnterFieldLabel(c *FieldLabelContext)

	// EnterField is called when entering the field production.
	EnterField(c *FieldContext)

	// EnterFieldOptions is called when entering the fieldOptions production.
	EnterFieldOptions(c *FieldOptionsContext)

	// EnterFieldOption is called when entering the fieldOption production.
	EnterFieldOption(c *FieldOptionContext)

	// EnterFieldNumber is called when entering the fieldNumber production.
	EnterFieldNumber(c *FieldNumberContext)

	// EnterOneof is called when entering the oneof production.
	EnterOneof(c *OneofContext)

	// EnterOneofField is called when entering the oneofField production.
	EnterOneofField(c *OneofFieldContext)

	// EnterMapField is called when entering the mapField production.
	EnterMapField(c *MapFieldContext)

	// EnterKeyType is called when entering the keyType production.
	EnterKeyType(c *KeyTypeContext)

	// EnterFieldType is called when entering the fieldType production.
	EnterFieldType(c *FieldTypeContext)

	// EnterReserved is called when entering the reserved production.
	EnterReserved(c *ReservedContext)

	// EnterExtensions is called when entering the extensions production.
	EnterExtensions(c *ExtensionsContext)

	// EnterRanges is called when entering the ranges production.
	EnterRanges(c *RangesContext)

	// EnterOneRange is called when entering the oneRange production.
	EnterOneRange(c *OneRangeContext)

	// EnterReservedFieldNames is called when entering the reservedFieldNames production.
	EnterReservedFieldNames(c *ReservedFieldNamesContext)

	// EnterTopLevelDef is called when entering the topLevelDef production.
	EnterTopLevelDef(c *TopLevelDefContext)

	// EnterEnumDef is called when entering the enumDef production.
	EnterEnumDef(c *EnumDefContext)

	// EnterEnumBody is called when entering the enumBody production.
	EnterEnumBody(c *EnumBodyContext)

	// EnterEnumElement is called when entering the enumElement production.
	EnterEnumElement(c *EnumElementContext)

	// EnterEnumField is called when entering the enumField production.
	EnterEnumField(c *EnumFieldContext)

	// EnterEnumValueOptions is called when entering the enumValueOptions production.
	EnterEnumValueOptions(c *EnumValueOptionsContext)

	// EnterEnumValueOption is called when entering the enumValueOption production.
	EnterEnumValueOption(c *EnumValueOptionContext)

	// EnterMessageDef is called when entering the messageDef production.
	EnterMessageDef(c *MessageDefContext)

	// EnterMessageBody is called when entering the messageBody production.
	EnterMessageBody(c *MessageBodyContext)

	// EnterMessageElement is called when entering the messageElement production.
	EnterMessageElement(c *MessageElementContext)

	// EnterExtendDef is called when entering the extendDef production.
	EnterExtendDef(c *ExtendDefContext)

	// EnterServiceDef is called when entering the serviceDef production.
	EnterServiceDef(c *ServiceDefContext)

	// EnterServiceElement is called when entering the serviceElement production.
	EnterServiceElement(c *ServiceElementContext)

	// EnterRpc is called when entering the rpc production.
	EnterRpc(c *RpcContext)

	// EnterConstant is called when entering the constant production.
	EnterConstant(c *ConstantContext)

	// EnterBlockLit is called when entering the blockLit production.
	EnterBlockLit(c *BlockLitContext)

	// EnterEmptyStatement is called when entering the emptyStatement production.
	EnterEmptyStatement(c *EmptyStatementContext)

	// EnterIdent is called when entering the ident production.
	EnterIdent(c *IdentContext)

	// EnterFullIdent is called when entering the fullIdent production.
	EnterFullIdent(c *FullIdentContext)

	// EnterMessageName is called when entering the messageName production.
	EnterMessageName(c *MessageNameContext)

	// EnterEnumName is called when entering the enumName production.
	EnterEnumName(c *EnumNameContext)

	// EnterFieldName is called when entering the fieldName production.
	EnterFieldName(c *FieldNameContext)

	// EnterOneofName is called when entering the oneofName production.
	EnterOneofName(c *OneofNameContext)

	// EnterServiceName is called when entering the serviceName production.
	EnterServiceName(c *ServiceNameContext)

	// EnterRpcName is called when entering the rpcName production.
	EnterRpcName(c *RpcNameContext)

	// EnterMessageType is called when entering the messageType production.
	EnterMessageType(c *MessageTypeContext)

	// EnterEnumType is called when entering the enumType production.
	EnterEnumType(c *EnumTypeContext)

	// EnterIntLit is called when entering the intLit production.
	EnterIntLit(c *IntLitContext)

	// EnterStrLit is called when entering the strLit production.
	EnterStrLit(c *StrLitContext)

	// EnterBoolLit is called when entering the boolLit production.
	EnterBoolLit(c *BoolLitContext)

	// EnterFloatLit is called when entering the floatLit production.
	EnterFloatLit(c *FloatLitContext)

	// EnterKeywords is called when entering the keywords production.
	EnterKeywords(c *KeywordsContext)

	// ExitProto is called when exiting the proto production.
	ExitProto(c *ProtoContext)

	// ExitEdition is called when exiting the edition production.
	ExitEdition(c *EditionContext)

	// ExitImportStatement is called when exiting the importStatement production.
	ExitImportStatement(c *ImportStatementContext)

	// ExitPackageStatement is called when exiting the packageStatement production.
	ExitPackageStatement(c *PackageStatementContext)

	// ExitOptionStatement is called when exiting the optionStatement production.
	ExitOptionStatement(c *OptionStatementContext)

	// ExitOptionName is called when exiting the optionName production.
	ExitOptionName(c *OptionNameContext)

	// ExitFieldLabel is called when exiting the fieldLabel production.
	ExitFieldLabel(c *FieldLabelContext)

	// ExitField is called when exiting the field production.
	ExitField(c *FieldContext)

	// ExitFieldOptions is called when exiting the fieldOptions production.
	ExitFieldOptions(c *FieldOptionsContext)

	// ExitFieldOption is called when exiting the fieldOption production.
	ExitFieldOption(c *FieldOptionContext)

	// ExitFieldNumber is called when exiting the fieldNumber production.
	ExitFieldNumber(c *FieldNumberContext)

	// ExitOneof is called when exiting the oneof production.
	ExitOneof(c *OneofContext)

	// ExitOneofField is called when exiting the oneofField production.
	ExitOneofField(c *OneofFieldContext)

	// ExitMapField is called when exiting the mapField production.
	ExitMapField(c *MapFieldContext)

	// ExitKeyType is called when exiting the keyType production.
	ExitKeyType(c *KeyTypeContext)

	// ExitFieldType is called when exiting the fieldType production.
	ExitFieldType(c *FieldTypeContext)

	// ExitReserved is called when exiting the reserved production.
	ExitReserved(c *ReservedContext)

	// ExitExtensions is called when exiting the extensions production.
	ExitExtensions(c *ExtensionsContext)

	// ExitRanges is called when exiting the ranges production.
	ExitRanges(c *RangesContext)

	// ExitOneRange is called when exiting the oneRange production.
	ExitOneRange(c *OneRangeContext)

	// ExitReservedFieldNames is called when exiting the reservedFieldNames production.
	ExitReservedFieldNames(c *ReservedFieldNamesContext)

	// ExitTopLevelDef is called when exiting the topLevelDef production.
	ExitTopLevelDef(c *TopLevelDefContext)

	// ExitEnumDef is called when exiting the enumDef production.
	ExitEnumDef(c *EnumDefContext)

	// ExitEnumBody is called when exiting the enumBody production.
	ExitEnumBody(c *EnumBodyContext)

	// ExitEnumElement is called when exiting the enumElement production.
	ExitEnumElement(c *EnumElementContext)

	// ExitEnumField is called when exiting the enumField production.
	ExitEnumField(c *EnumFieldContext)

	// ExitEnumValueOptions is called when exiting the enumValueOptions production.
	ExitEnumValueOptions(c *EnumValueOptionsContext)

	// ExitEnumValueOption is called when exiting the enumValueOption production.
	ExitEnumValueOption(c *EnumValueOptionContext)

	// ExitMessageDef is called when exiting the messageDef production.
	ExitMessageDef(c *MessageDefContext)

	// ExitMessageBody is called when exiting the messageBody production.
	ExitMessageBody(c *MessageBodyContext)

	// ExitMessageElement is called when exiting the messageElement production.
	ExitMessageElement(c *MessageElementContext)

	// ExitExtendDef is called when exiting the extendDef production.
	ExitExtendDef(c *ExtendDefContext)

	// ExitServiceDef is called when exiting the serviceDef production.
	ExitServiceDef(c *ServiceDefContext)

	// ExitServiceElement is called when exiting the serviceElement production.
	ExitServiceElement(c *ServiceElementContext)

	// ExitRpc is called when exiting the rpc production.
	ExitRpc(c *RpcContext)

	// ExitConstant is called when exiting the constant production.
	ExitConstant(c *ConstantContext)

	// ExitBlockLit is called when exiting the blockLit production.
	ExitBlockLit(c *BlockLitContext)

	// ExitEmptyStatement is called when exiting the emptyStatement production.
	ExitEmptyStatement(c *EmptyStatementContext)

	// ExitIdent is called when exiting the ident production.
	ExitIdent(c *IdentContext)

	// ExitFullIdent is called when exiting the fullIdent production.
	ExitFullIdent(c *FullIdentContext)

	// ExitMessageName is called when exiting the messageName production.
	ExitMessageName(c *MessageNameContext)

	// ExitEnumName is called when exiting the enumName production.
	ExitEnumName(c *EnumNameContext)

	// ExitFieldName is called when exiting the fieldName production.
	ExitFieldName(c *FieldNameContext)

	// ExitOneofName is called when exiting the oneofName production.
	ExitOneofName(c *OneofNameContext)

	// ExitServiceName is called when exiting the serviceName production.
	ExitServiceName(c *ServiceNameContext)

	// ExitRpcName is called when exiting the rpcName production.
	ExitRpcName(c *RpcNameContext)

	// ExitMessageType is called when exiting the messageType production.
	ExitMessageType(c *MessageTypeContext)

	// ExitEnumType is called when exiting the enumType production.
	ExitEnumType(c *EnumTypeContext)

	// ExitIntLit is called when exiting the intLit production.
	ExitIntLit(c *IntLitContext)

	// ExitStrLit is called when exiting the strLit production.
	ExitStrLit(c *StrLitContext)

	// ExitBoolLit is called when exiting the boolLit production.
	ExitBoolLit(c *BoolLitContext)

	// ExitFloatLit is called when exiting the floatLit production.
	ExitFloatLit(c *FloatLitContext)

	// ExitKeywords is called when exiting the keywords production.
	ExitKeywords(c *KeywordsContext)
}
