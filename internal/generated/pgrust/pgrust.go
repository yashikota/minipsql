package pgrust

import (
	base "github.com/yashikota/minipsql/internal/generated/pgrust/base"
	"unsafe"
	_ "github.com/yashikota/minipsql/internal/generated/pgrust/p32"
	_ "embed"
)

func New(env base.EnvImports, pgvfs base.PgvfsImports) *base.Module {
	m := &base.Module{Env: env, Pgvfs: pgvfs}
	m.Memory = make([]byte, 8781824, 10977280)
	m.M = unsafe.Pointer(unsafe.SliceData(m.Memory))
	m.MaxMem = 4294967296
	m.T0 = make([]any, 21841)
	m.G0 = int64(1048576)
	InitElemSeg_0_0(m)
	InitElemSeg_1_0(m)
	InitElemSeg_2_0(m)
	InitElemSeg_3_0(m)
	InitElemSeg_4_0(m)
	InitElemSeg_5_0(m)
	InitElemSeg_6_0(m)
	InitElemSeg_7_0(m)
	InitElemSeg_8_0(m)
	InitElemSeg_9_0(m)
	InitElemSeg_10_0(m)
	InitElemSeg_11_0(m)
	InitElemSeg_12_0(m)
	InitElemSeg_13_0(m)
	InitElemSeg_14_0(m)
	InitElemSeg_15_0(m)
	InitElemSeg_16_0(m)
	InitElemSeg_17_0(m)
	InitElemSeg_18_0(m)
	InitElemSeg_19_0(m)
	InitElemSeg_20_0(m)
	InitElemSeg_21_0(m)
	InitElemSeg_22_0(m)
	InitElemSeg_22_1(m)
	InitElemSeg_23_0(m)
	InitElemSeg_23_1(m)
	InitElemSeg_24_0(m)
	InitElemSeg_24_1(m)
	InitElemSeg_25_0(m)
	InitElemSeg_25_1(m)
	InitElemSeg_25_2(m)
	InitElemSeg_26_0(m)
	InitElemSeg_26_1(m)
	InitElemSeg_26_2(m)
	InitElemSeg_27_0(m)
	InitElemSeg_27_1(m)
	InitElemSeg_27_2(m)
	InitElemSeg_27_3(m)
	InitElemSeg_28_0(m)
	InitElemSeg_28_1(m)
	InitElemSeg_28_2(m)
	InitElemSeg_28_3(m)
	InitElemSeg_29_0(m)
	InitElemSeg_29_1(m)
	InitElemSeg_29_2(m)
	InitElemSeg_29_3(m)
	InitElemSeg_29_4(m)
	InitElemSeg_29_5(m)
	InitElemSeg_30_0(m)
	InitElemSeg_30_1(m)
	InitElemSeg_30_2(m)
	InitElemSeg_30_3(m)
	InitElemSeg_30_4(m)
	InitElemSeg_30_5(m)
	InitElemSeg_30_6(m)
	InitElemSeg_30_7(m)
	InitElemSeg_31_0(m)
	InitElemSeg_31_1(m)
	InitElemSeg_31_2(m)
	InitElemSeg_31_3(m)
	InitElemSeg_31_4(m)
	InitElemSeg_31_5(m)
	InitElemSeg_31_6(m)
	InitElemSeg_31_7(m)
	InitElemSeg_31_8(m)
	InitElemSeg_31_9(m)
	InitElemSeg_31_10(m)
	InitElemSeg_31_11(m)
	InitElemSeg_31_12(m)
	InitElemSeg_31_13(m)
	InitElemSeg_31_14(m)
	InitElemSeg_31_15(m)
	InitElemSeg_32_0(m)
	InitElemSeg_32_1(m)
	InitElemSeg_32_2(m)
	InitElemSeg_32_3(m)
	InitElemSeg_32_4(m)
	InitElemSeg_32_5(m)
	InitElemSeg_32_6(m)
	InitElemSeg_32_7(m)
	InitElemSeg_32_8(m)
	InitElemSeg_32_9(m)
	InitElemSeg_32_10(m)
	InitElemSeg_32_11(m)
	InitElemSeg_32_12(m)
	InitElemSeg_32_13(m)
	InitElemSeg_32_14(m)
	InitElemSeg_32_15(m)
	InitElemSeg_32_16(m)
	InitElemSeg_32_17(m)
	InitElemSeg_32_18(m)
	InitElemSeg_32_19(m)
	InitElemSeg_32_20(m)
	InitElemSeg_32_21(m)
	InitElemSeg_32_22(m)
	InitElemSeg_32_23(m)
	InitElemSeg_32_24(m)
	InitElemSeg_32_25(m)
	InitElemSeg_32_26(m)
	InitElemSeg_32_27(m)
	InitElemSeg_32_28(m)
	InitElemSeg_32_29(m)
	InitElemSeg_32_30(m)
	InitElemSeg_32_31(m)
	InitElemSeg_32_32(m)
	InitElemSeg_32_33(m)
	InitElemSeg_32_34(m)
	InitElemSeg_32_35(m)
	initData_0(m)
	return m
}
func initData_0(m *base.Module) {
	copy(m.Memory[1048576:], wasm2goData_data_bin[0:4230253])
	copy(m.Memory[5287024:], wasm2goData_data_bin[4230253:6870000])
	copy(m.Memory[7930040:], wasm2goData_data_bin[6870000:6905712])
	copy(m.Memory[7966856:], wasm2goData_data_bin[6905712:6905844])
	copy(m.Memory[7968460:], wasm2goData_data_bin[6905844:7299465])
	copy(m.Memory[8493360:], wasm2goData_data_bin[7299465:7299634])
	copy(m.Memory[8494629:], wasm2goData_data_bin[7299634:7300181])
}
func GetrandomCustom(m *base.Module, l0 int64, l1 int64) int32 {
	return Fn34103(m, l0, l1)
}
func SystemFuncName(m *base.Module, l0 int64) int64 {
	return Fn14602(m, l0)
}
func SystemTypeName(m *base.Module, l0 int64) int64 {
	return Fn14603(m, l0)
}
func BaseYyparse(m *base.Module, l0 int64) int32 {
	return Fn14620(m, l0)
}
func AllocSetAlloc(m *base.Module, l0 int64, l1 int64, l2 int32) int64 {
	return Fn20470(m, l0, l1, l2)
}
func AllocSetDelete(m *base.Module, l0 int64) {
	Fn20475(m, l0)
}
func AllocSetFree(m *base.Module, l0 int64) {
	Fn20477(m, l0)
}
func AllocSetGetChunkContext(m *base.Module, l0 int64) int64 {
	return Fn20479(m, l0)
}
func AllocSetGetChunkSpace(m *base.Module, l0 int64) int64 {
	return Fn20480(m, l0)
}
func AllocSetIsEmpty(m *base.Module, l0 int64) int32 {
	return Fn20481(m, l0)
}
func AllocSetRealloc(m *base.Module, l0 int64, l1 int64, l2 int32) int64 {
	return Fn20482(m, l0, l1, l2)
}
func AllocSetReset(m *base.Module, l0 int64) {
	Fn20484(m, l0)
}
func AllocSetStats(m *base.Module, l0 int64, l1 int64, l2 int64, l3 int64, l4 int32) {
	Fn20486(m, l0, l1, l2, l3, l4)
}
func Main(m *base.Module, l0 int32, l1 int64) int32 {
	return Fn34104(m, l0, l1)
}
func Memory(m *base.Module) []byte {
	return m.Memory
}
//go:embed data.bin
var wasm2goData_data_bin []byte
