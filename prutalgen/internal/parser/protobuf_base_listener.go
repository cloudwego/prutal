// Code generated from ./Protobuf.g4 by ANTLR 4.13.2. DO NOT EDIT.

package parser // Protobuf

import "github.com/cloudwego/prutal/prutalgen/internal/antlr"

// BaseProtobufListener is a complete listener for a parse tree produced by ProtobufParser.
type BaseProtobufListener struct{}

var _ ProtobufListener = &BaseProtobufListener{}

// VisitTerminal is called when a terminal node is visited.
func (s *BaseProtobufListener) VisitTerminal(node antlr.TerminalNode) {}

// VisitErrorNode is called when an error node is visited.
func (s *BaseProtobufListener) VisitErrorNode(node antlr.ErrorNode) {}

// EnterEveryRule is called when any rule is entered.
func (s *BaseProtobufListener) EnterEveryRule(ctx antlr.ParserRuleContext) {}

// ExitEveryRule is called when any rule is exited.
func (s *BaseProtobufListener) ExitEveryRule(ctx antlr.ParserRuleContext) {}

// EnterProto is called when production proto is entered.
func (s *BaseProtobufListener) EnterProto(ctx *ProtoContext) {}

// ExitProto is called when production proto is exited.
func (s *BaseProtobufListener) ExitProto(ctx *ProtoContext) {}

// EnterEdition is called when production edition is entered.
func (s *BaseProtobufListener) EnterEdition(ctx *EditionContext) {}

// ExitEdition is called when production edition is exited.
func (s *BaseProtobufListener) ExitEdition(ctx *EditionContext) {}

// EnterImportStatement is called when production importStatement is entered.
func (s *BaseProtobufListener) EnterImportStatement(ctx *ImportStatementContext) {}

// ExitImportStatement is called when production importStatement is exited.
func (s *BaseProtobufListener) ExitImportStatement(ctx *ImportStatementContext) {}

// EnterPackageStatement is called when production packageStatement is entered.
func (s *BaseProtobufListener) EnterPackageStatement(ctx *PackageStatementContext) {}

// ExitPackageStatement is called when production packageStatement is exited.
func (s *BaseProtobufListener) ExitPackageStatement(ctx *PackageStatementContext) {}

// EnterOptionStatement is called when production optionStatement is entered.
func (s *BaseProtobufListener) EnterOptionStatement(ctx *OptionStatementContext) {}

// ExitOptionStatement is called when production optionStatement is exited.
func (s *BaseProtobufListener) ExitOptionStatement(ctx *OptionStatementContext) {}

// EnterOptionName is called when production optionName is entered.
func (s *BaseProtobufListener) EnterOptionName(ctx *OptionNameContext) {}

// ExitOptionName is called when production optionName is exited.
func (s *BaseProtobufListener) ExitOptionName(ctx *OptionNameContext) {}

// EnterFieldLabel is called when production fieldLabel is entered.
func (s *BaseProtobufListener) EnterFieldLabel(ctx *FieldLabelContext) {}

// ExitFieldLabel is called when production fieldLabel is exited.
func (s *BaseProtobufListener) ExitFieldLabel(ctx *FieldLabelContext) {}

// EnterField is called when production field is entered.
func (s *BaseProtobufListener) EnterField(ctx *FieldContext) {}

// ExitField is called when production field is exited.
func (s *BaseProtobufListener) ExitField(ctx *FieldContext) {}

// EnterFieldOptions is called when production fieldOptions is entered.
func (s *BaseProtobufListener) EnterFieldOptions(ctx *FieldOptionsContext) {}

// ExitFieldOptions is called when production fieldOptions is exited.
func (s *BaseProtobufListener) ExitFieldOptions(ctx *FieldOptionsContext) {}

// EnterFieldOption is called when production fieldOption is entered.
func (s *BaseProtobufListener) EnterFieldOption(ctx *FieldOptionContext) {}

// ExitFieldOption is called when production fieldOption is exited.
func (s *BaseProtobufListener) ExitFieldOption(ctx *FieldOptionContext) {}

// EnterFieldNumber is called when production fieldNumber is entered.
func (s *BaseProtobufListener) EnterFieldNumber(ctx *FieldNumberContext) {}

// ExitFieldNumber is called when production fieldNumber is exited.
func (s *BaseProtobufListener) ExitFieldNumber(ctx *FieldNumberContext) {}

// EnterOneof is called when production oneof is entered.
func (s *BaseProtobufListener) EnterOneof(ctx *OneofContext) {}

// ExitOneof is called when production oneof is exited.
func (s *BaseProtobufListener) ExitOneof(ctx *OneofContext) {}

// EnterOneofField is called when production oneofField is entered.
func (s *BaseProtobufListener) EnterOneofField(ctx *OneofFieldContext) {}

// ExitOneofField is called when production oneofField is exited.
func (s *BaseProtobufListener) ExitOneofField(ctx *OneofFieldContext) {}

// EnterMapField is called when production mapField is entered.
func (s *BaseProtobufListener) EnterMapField(ctx *MapFieldContext) {}

// ExitMapField is called when production mapField is exited.
func (s *BaseProtobufListener) ExitMapField(ctx *MapFieldContext) {}

// EnterKeyType is called when production keyType is entered.
func (s *BaseProtobufListener) EnterKeyType(ctx *KeyTypeContext) {}

// ExitKeyType is called when production keyType is exited.
func (s *BaseProtobufListener) ExitKeyType(ctx *KeyTypeContext) {}

// EnterFieldType is called when production fieldType is entered.
func (s *BaseProtobufListener) EnterFieldType(ctx *FieldTypeContext) {}

// ExitFieldType is called when production fieldType is exited.
func (s *BaseProtobufListener) ExitFieldType(ctx *FieldTypeContext) {}

// EnterReserved is called when production reserved is entered.
func (s *BaseProtobufListener) EnterReserved(ctx *ReservedContext) {}

// ExitReserved is called when production reserved is exited.
func (s *BaseProtobufListener) ExitReserved(ctx *ReservedContext) {}

// EnterExtensions is called when production extensions is entered.
func (s *BaseProtobufListener) EnterExtensions(ctx *ExtensionsContext) {}

// ExitExtensions is called when production extensions is exited.
func (s *BaseProtobufListener) ExitExtensions(ctx *ExtensionsContext) {}

// EnterRanges is called when production ranges is entered.
func (s *BaseProtobufListener) EnterRanges(ctx *RangesContext) {}

// ExitRanges is called when production ranges is exited.
func (s *BaseProtobufListener) ExitRanges(ctx *RangesContext) {}

// EnterOneRange is called when production oneRange is entered.
func (s *BaseProtobufListener) EnterOneRange(ctx *OneRangeContext) {}

// ExitOneRange is called when production oneRange is exited.
func (s *BaseProtobufListener) ExitOneRange(ctx *OneRangeContext) {}

// EnterReservedFieldNames is called when production reservedFieldNames is entered.
func (s *BaseProtobufListener) EnterReservedFieldNames(ctx *ReservedFieldNamesContext) {}

// ExitReservedFieldNames is called when production reservedFieldNames is exited.
func (s *BaseProtobufListener) ExitReservedFieldNames(ctx *ReservedFieldNamesContext) {}

// EnterTopLevelDef is called when production topLevelDef is entered.
func (s *BaseProtobufListener) EnterTopLevelDef(ctx *TopLevelDefContext) {}

// ExitTopLevelDef is called when production topLevelDef is exited.
func (s *BaseProtobufListener) ExitTopLevelDef(ctx *TopLevelDefContext) {}

// EnterEnumDef is called when production enumDef is entered.
func (s *BaseProtobufListener) EnterEnumDef(ctx *EnumDefContext) {}

// ExitEnumDef is called when production enumDef is exited.
func (s *BaseProtobufListener) ExitEnumDef(ctx *EnumDefContext) {}

// EnterEnumBody is called when production enumBody is entered.
func (s *BaseProtobufListener) EnterEnumBody(ctx *EnumBodyContext) {}

// ExitEnumBody is called when production enumBody is exited.
func (s *BaseProtobufListener) ExitEnumBody(ctx *EnumBodyContext) {}

// EnterEnumElement is called when production enumElement is entered.
func (s *BaseProtobufListener) EnterEnumElement(ctx *EnumElementContext) {}

// ExitEnumElement is called when production enumElement is exited.
func (s *BaseProtobufListener) ExitEnumElement(ctx *EnumElementContext) {}

// EnterEnumField is called when production enumField is entered.
func (s *BaseProtobufListener) EnterEnumField(ctx *EnumFieldContext) {}

// ExitEnumField is called when production enumField is exited.
func (s *BaseProtobufListener) ExitEnumField(ctx *EnumFieldContext) {}

// EnterEnumValueOptions is called when production enumValueOptions is entered.
func (s *BaseProtobufListener) EnterEnumValueOptions(ctx *EnumValueOptionsContext) {}

// ExitEnumValueOptions is called when production enumValueOptions is exited.
func (s *BaseProtobufListener) ExitEnumValueOptions(ctx *EnumValueOptionsContext) {}

// EnterEnumValueOption is called when production enumValueOption is entered.
func (s *BaseProtobufListener) EnterEnumValueOption(ctx *EnumValueOptionContext) {}

// ExitEnumValueOption is called when production enumValueOption is exited.
func (s *BaseProtobufListener) ExitEnumValueOption(ctx *EnumValueOptionContext) {}

// EnterMessageDef is called when production messageDef is entered.
func (s *BaseProtobufListener) EnterMessageDef(ctx *MessageDefContext) {}

// ExitMessageDef is called when production messageDef is exited.
func (s *BaseProtobufListener) ExitMessageDef(ctx *MessageDefContext) {}

// EnterMessageBody is called when production messageBody is entered.
func (s *BaseProtobufListener) EnterMessageBody(ctx *MessageBodyContext) {}

// ExitMessageBody is called when production messageBody is exited.
func (s *BaseProtobufListener) ExitMessageBody(ctx *MessageBodyContext) {}

// EnterMessageElement is called when production messageElement is entered.
func (s *BaseProtobufListener) EnterMessageElement(ctx *MessageElementContext) {}

// ExitMessageElement is called when production messageElement is exited.
func (s *BaseProtobufListener) ExitMessageElement(ctx *MessageElementContext) {}

// EnterExtendDef is called when production extendDef is entered.
func (s *BaseProtobufListener) EnterExtendDef(ctx *ExtendDefContext) {}

// ExitExtendDef is called when production extendDef is exited.
func (s *BaseProtobufListener) ExitExtendDef(ctx *ExtendDefContext) {}

// EnterServiceDef is called when production serviceDef is entered.
func (s *BaseProtobufListener) EnterServiceDef(ctx *ServiceDefContext) {}

// ExitServiceDef is called when production serviceDef is exited.
func (s *BaseProtobufListener) ExitServiceDef(ctx *ServiceDefContext) {}

// EnterServiceElement is called when production serviceElement is entered.
func (s *BaseProtobufListener) EnterServiceElement(ctx *ServiceElementContext) {}

// ExitServiceElement is called when production serviceElement is exited.
func (s *BaseProtobufListener) ExitServiceElement(ctx *ServiceElementContext) {}

// EnterRpc is called when production rpc is entered.
func (s *BaseProtobufListener) EnterRpc(ctx *RpcContext) {}

// ExitRpc is called when production rpc is exited.
func (s *BaseProtobufListener) ExitRpc(ctx *RpcContext) {}

// EnterConstant is called when production constant is entered.
func (s *BaseProtobufListener) EnterConstant(ctx *ConstantContext) {}

// ExitConstant is called when production constant is exited.
func (s *BaseProtobufListener) ExitConstant(ctx *ConstantContext) {}

// EnterBlockLit is called when production blockLit is entered.
func (s *BaseProtobufListener) EnterBlockLit(ctx *BlockLitContext) {}

// ExitBlockLit is called when production blockLit is exited.
func (s *BaseProtobufListener) ExitBlockLit(ctx *BlockLitContext) {}

// EnterEmptyStatement is called when production emptyStatement is entered.
func (s *BaseProtobufListener) EnterEmptyStatement(ctx *EmptyStatementContext) {}

// ExitEmptyStatement is called when production emptyStatement is exited.
func (s *BaseProtobufListener) ExitEmptyStatement(ctx *EmptyStatementContext) {}

// EnterIdent is called when production ident is entered.
func (s *BaseProtobufListener) EnterIdent(ctx *IdentContext) {}

// ExitIdent is called when production ident is exited.
func (s *BaseProtobufListener) ExitIdent(ctx *IdentContext) {}

// EnterFullIdent is called when production fullIdent is entered.
func (s *BaseProtobufListener) EnterFullIdent(ctx *FullIdentContext) {}

// ExitFullIdent is called when production fullIdent is exited.
func (s *BaseProtobufListener) ExitFullIdent(ctx *FullIdentContext) {}

// EnterMessageName is called when production messageName is entered.
func (s *BaseProtobufListener) EnterMessageName(ctx *MessageNameContext) {}

// ExitMessageName is called when production messageName is exited.
func (s *BaseProtobufListener) ExitMessageName(ctx *MessageNameContext) {}

// EnterEnumName is called when production enumName is entered.
func (s *BaseProtobufListener) EnterEnumName(ctx *EnumNameContext) {}

// ExitEnumName is called when production enumName is exited.
func (s *BaseProtobufListener) ExitEnumName(ctx *EnumNameContext) {}

// EnterFieldName is called when production fieldName is entered.
func (s *BaseProtobufListener) EnterFieldName(ctx *FieldNameContext) {}

// ExitFieldName is called when production fieldName is exited.
func (s *BaseProtobufListener) ExitFieldName(ctx *FieldNameContext) {}

// EnterOneofName is called when production oneofName is entered.
func (s *BaseProtobufListener) EnterOneofName(ctx *OneofNameContext) {}

// ExitOneofName is called when production oneofName is exited.
func (s *BaseProtobufListener) ExitOneofName(ctx *OneofNameContext) {}

// EnterServiceName is called when production serviceName is entered.
func (s *BaseProtobufListener) EnterServiceName(ctx *ServiceNameContext) {}

// ExitServiceName is called when production serviceName is exited.
func (s *BaseProtobufListener) ExitServiceName(ctx *ServiceNameContext) {}

// EnterRpcName is called when production rpcName is entered.
func (s *BaseProtobufListener) EnterRpcName(ctx *RpcNameContext) {}

// ExitRpcName is called when production rpcName is exited.
func (s *BaseProtobufListener) ExitRpcName(ctx *RpcNameContext) {}

// EnterMessageType is called when production messageType is entered.
func (s *BaseProtobufListener) EnterMessageType(ctx *MessageTypeContext) {}

// ExitMessageType is called when production messageType is exited.
func (s *BaseProtobufListener) ExitMessageType(ctx *MessageTypeContext) {}

// EnterEnumType is called when production enumType is entered.
func (s *BaseProtobufListener) EnterEnumType(ctx *EnumTypeContext) {}

// ExitEnumType is called when production enumType is exited.
func (s *BaseProtobufListener) ExitEnumType(ctx *EnumTypeContext) {}

// EnterIntLit is called when production intLit is entered.
func (s *BaseProtobufListener) EnterIntLit(ctx *IntLitContext) {}

// ExitIntLit is called when production intLit is exited.
func (s *BaseProtobufListener) ExitIntLit(ctx *IntLitContext) {}

// EnterStrLit is called when production strLit is entered.
func (s *BaseProtobufListener) EnterStrLit(ctx *StrLitContext) {}

// ExitStrLit is called when production strLit is exited.
func (s *BaseProtobufListener) ExitStrLit(ctx *StrLitContext) {}

// EnterBoolLit is called when production boolLit is entered.
func (s *BaseProtobufListener) EnterBoolLit(ctx *BoolLitContext) {}

// ExitBoolLit is called when production boolLit is exited.
func (s *BaseProtobufListener) ExitBoolLit(ctx *BoolLitContext) {}

// EnterFloatLit is called when production floatLit is entered.
func (s *BaseProtobufListener) EnterFloatLit(ctx *FloatLitContext) {}

// ExitFloatLit is called when production floatLit is exited.
func (s *BaseProtobufListener) ExitFloatLit(ctx *FloatLitContext) {}

// EnterKeywords is called when production keywords is entered.
func (s *BaseProtobufListener) EnterKeywords(ctx *KeywordsContext) {}

// ExitKeywords is called when production keywords is exited.
func (s *BaseProtobufListener) ExitKeywords(ctx *KeywordsContext) {}
