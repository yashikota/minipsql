package base

import (
	"math"
	"math/bits"
	"runtime"
	"sync"
	"unsafe"
)

type EnvImports interface {
	BIO_get_data(m *Module, l0 int64) int64
	BIO_clear_flags(m *Module, l0 int64, l1 int32)
	BIO_s_mem(m *Module) int64
	BIO_new(m *Module, l0 int64) int64
	X509_NAME_entry_count(m *Module, l0 int64) int32
	X509_NAME_get_entry(m *Module, l0 int64, l1 int32) int64
	X509_NAME_ENTRY_get_object(m *Module, l0 int64) int64
	OBJ_obj2nid(m *Module, l0 int64) int32
	X509_NAME_ENTRY_get_data(m *Module, l0 int64) int64
	OBJ_nid2sn(m *Module, l0 int32) int64
	OBJ_nid2ln(m *Module, l0 int32) int64
	BIO_printf(m *Module, l0 int64, l1 int64, l2 int64) int32
	ASN1_STRING_print_ex(m *Module, l0 int64, l1 int64, l2 int64) int32
	BIO_write(m *Module, l0 int64, l1 int64, l2 int32) int32
	BIO_ctrl(m *Module, l0 int64, l1 int32, l2 int64, l3 int64) int64
	BIO_free(m *Module, l0 int64) int32
	SSL_select_next_proto(m *Module, l0 int64, l1 int64, l2 int64, l3 int32, l4 int64, l5 int32) int32
	SSL_state_string_long(m *Module, l0 int64) int64
	X509_STORE_CTX_get_error_depth(m *Module, l0 int64) int32
	X509_STORE_CTX_get_error(m *Module, l0 int64) int32
	X509_verify_cert_error_string(m *Module, l0 int64) int64
	X509_STORE_CTX_get_current_cert(m *Module, l0 int64) int64
	X509_get_subject_name(m *Module, l0 int64) int64
	X509_get_issuer_name(m *Module, l0 int64) int64
	X509_get_serialNumber(m *Module, l0 int64) int64
	ASN1_INTEGER_to_BN(m *Module, l0 int64, l1 int64) int64
	BN_bn2dec(m *Module, l0 int64) int64
	CRYPTO_free(m *Module, l0 int64, l1 int64, l2 int32)
	BN_free(m *Module, l0 int64)
	TLS_method(m *Module) int64
	SSL_CTX_new(m *Module, l0 int64) int64
	SSL_CTX_ctrl(m *Module, l0 int64, l1 int32, l2 int64, l3 int64) int64
	SSL_CTX_set_default_passwd_cb(m *Module, l0 int64, l1 int64)
	SSL_CTX_use_certificate_chain_file(m *Module, l0 int64, l1 int64) int32
	SSL_CTX_use_PrivateKey_file(m *Module, l0 int64, l1 int64, l2 int32) int32
	SSL_CTX_check_private_key(m *Module, l0 int64) int32
	SSL_CTX_set_options(m *Module, l0 int64, l1 int64) int64
	X509_free(m *Module, l0 int64)
	SSL_shutdown(m *Module, l0 int64) int32
	SSL_free(m *Module, l0 int64)
	ERR_clear_error(m *Module)
	SSL_read(m *Module, l0 int64, l1 int64, l2 int32) int32
	SSL_get_error(m *Module, l0 int64, l1 int32) int32
	ERR_get_error(m *Module) int64
	SSL_write(m *Module, l0 int64, l1 int64, l2 int32) int32
	ERR_reason_error_string(m *Module, l0 int64) int64
	SSL_get_version(m *Module, l0 int64) int64
	SSL_get_current_cipher(m *Module, l0 int64) int64
	SSL_CIPHER_get_name(m *Module, l0 int64) int64
	SSL_CIPHER_get_bits(m *Module, l0 int64, l1 int64) int32
	SSL_CTX_free(m *Module, l0 int64)
	Fopen(m *Module, l0 int64, l1 int64) int64
	PEM_read_DHparams(m *Module, l0 int64, l1 int64, l2 int64, l3 int64) int64
	Fclose(m *Module, l0 int64) int32
	BIO_new_mem_buf(m *Module, l0 int64, l1 int32) int64
	PEM_read_bio_DHparams(m *Module, l0 int64, l1 int64, l2 int64, l3 int64) int64
	DH_check(m *Module, l0 int64, l1 int64) int32
	DH_free(m *Module, l0 int64)
	SSL_CTX_set_cipher_list(m *Module, l0 int64, l1 int64) int32
	SSL_CTX_set_ciphersuites(m *Module, l0 int64, l1 int64) int32
	SSL_CTX_load_verify_locations(m *Module, l0 int64, l1 int64, l2 int64) int32
	SSL_load_client_CA_file(m *Module, l0 int64) int64
	SSL_CTX_set_client_CA_list(m *Module, l0 int64, l1 int64)
	SSL_CTX_set_verify(m *Module, l0 int64, l1 int32, l2 int64)
	SSL_CTX_get_cert_store(m *Module, l0 int64) int64
	X509_STORE_load_locations(m *Module, l0 int64, l1 int64, l2 int64) int32
	X509_STORE_set_flags(m *Module, l0 int64, l1 int64) int32
	SSL_CTX_set_info_callback(m *Module, l0 int64, l1 int64)
	SSL_CTX_set_alpn_select_cb(m *Module, l0 int64, l1 int64, l2 int64)
	SSL_new(m *Module, l0 int64) int64
	BIO_get_new_index(m *Module) int32
	BIO_meth_new(m *Module, l0 int32, l1 int64) int64
	BIO_meth_set_write(m *Module, l0 int64, l1 int64) int32
	BIO_meth_set_read(m *Module, l0 int64, l1 int64) int32
	BIO_meth_set_ctrl(m *Module, l0 int64, l1 int64) int32
	BIO_set_data(m *Module, l0 int64, l1 int64)
	BIO_set_init(m *Module, l0 int64, l1 int32)
	SSL_set_bio(m *Module, l0 int64, l1 int64, l2 int64)
	SSL_accept(m *Module, l0 int64) int32
	SSL_get0_alpn_selected(m *Module, l0 int64, l1 int64, l2 int64)
	SSL_get1_peer_certificate(m *Module, l0 int64) int64
	SSL_get_certificate(m *Module, l0 int64) int64
	X509_get_signature_info(m *Module, l0 int64, l1 int64, l2 int64, l3 int64, l4 int64) int32
	EVP_sha256(m *Module) int64
	EVP_get_digestbyname(m *Module, l0 int64) int64
	X509_digest(m *Module, l0 int64, l1 int64, l2 int64, l3 int64) int32
	X509_NAME_get_text_by_NID(m *Module, l0 int64, l1 int32, l2 int64, l3 int32) int32
	X509_NAME_print_ex(m *Module, l0 int64, l1 int64, l2 int32, l3 int64) int32
	Lgamma(m *Module, l0 float64) float64
	Mbstowcs(m *Module, l0 int64, l1 int64, l2 int64) int64
	Ucol_close_0(m *Module, l0 int64)
	Strcoll_l(m *Module, l0 int64, l1 int64, l2 int64) int32
	Strxfrm_l(m *Module, l0 int64, l1 int64, l2 int64, l3 int64) int64
	Tolower_l(m *Module, l0 int32, l1 int64) int32
	Iswdigit_l(m *Module, l0 int32, l1 int64) int32
	Iswalpha_l(m *Module, l0 int32, l1 int64) int32
	Iswalnum_l(m *Module, l0 int32, l1 int64) int32
	Iswupper_l(m *Module, l0 int32, l1 int64) int32
	Iswlower_l(m *Module, l0 int32, l1 int64) int32
	Iswgraph_l(m *Module, l0 int32, l1 int64) int32
	Iswprint_l(m *Module, l0 int32, l1 int64) int32
	Iswpunct_l(m *Module, l0 int32, l1 int64) int32
	Iswspace_l(m *Module, l0 int32, l1 int64) int32
	Isdigit_l(m *Module, l0 int32, l1 int64) int32
	Isalpha_l(m *Module, l0 int32, l1 int64) int32
	Isalnum_l(m *Module, l0 int32, l1 int64) int32
	Isupper_l(m *Module, l0 int32, l1 int64) int32
	Islower_l(m *Module, l0 int32, l1 int64) int32
	Isgraph_l(m *Module, l0 int32, l1 int64) int32
	Isprint_l(m *Module, l0 int32, l1 int64) int32
	Ispunct_l(m *Module, l0 int32, l1 int64) int32
	Isspace_l(m *Module, l0 int32, l1 int64) int32
	Towlower_l(m *Module, l0 int32, l1 int64) int32
	Toupper_l(m *Module, l0 int32, l1 int64) int32
	Towupper_l(m *Module, l0 int32, l1 int64) int32
	Nl_langinfo_l(m *Module, l0 int32, l1 int64) int64
	U_isdigit_0(m *Module, l0 int32) int32
	U_isalpha_0(m *Module, l0 int32) int32
	U_isalnum_0(m *Module, l0 int32) int32
	U_isupper_0(m *Module, l0 int32) int32
	U_islower_0(m *Module, l0 int32) int32
	U_isgraph_0(m *Module, l0 int32) int32
	U_isprint_0(m *Module, l0 int32) int32
	U_ispunct_0(m *Module, l0 int32) int32
	U_isspace_0(m *Module, l0 int32) int32
	U_tolower_0(m *Module, l0 int32) int32
	U_toupper_0(m *Module, l0 int32) int32
	U_strFoldCase_0(m *Module, l0 int64, l1 int32, l2 int64, l3 int32, l4 int32, l5 int64) int32
	Ucol_open_0(m *Module, l0 int64, l1 int64) int64
	U_strToLower_0(m *Module, l0 int64, l1 int32, l2 int64, l3 int32, l4 int64, l5 int64) int32
	Ucol_strcollUTF8_0(m *Module, l0 int64, l1 int64, l2 int32, l3 int64, l4 int32, l5 int64) int32
	Ucol_getSortKey_0(m *Module, l0 int64, l1 int64, l2 int32, l3 int64, l4 int32) int32
	U_strToTitle_0(m *Module, l0 int64, l1 int32, l2 int64, l3 int32, l4 int64, l5 int64, l6 int64) int32
	U_strToUpper_0(m *Module, l0 int64, l1 int32, l2 int64, l3 int32, l4 int64, l5 int64) int32
	U_strToUTF8_0(m *Module, l0 int64, l1 int32, l2 int64, l3 int64, l4 int32, l5 int64) int64
	U_strFromUTF8_0(m *Module, l0 int64, l1 int32, l2 int64, l3 int64, l4 int32, l5 int64) int64
	Ucol_getRules_0(m *Module, l0 int64, l1 int64) int64
	U_strlen_0(m *Module, l0 int64) int32
	Ucol_openRules_0(m *Module, l0 int64, l1 int32, l2 int32, l3 int32, l4 int64, l5 int64) int64
	Uloc_toLanguageTag_0(m *Module, l0 int64, l1 int64, l2 int32, l3 int32, l4 int64) int32
	Uloc_countAvailable_0(m *Module) int32
	Uloc_getAvailable_0(m *Module, l0 int32) int64
	Uloc_getLanguage_0(m *Module, l0 int64, l1 int64, l2 int32, l3 int64) int32
	Uloc_getDisplayName_0(m *Module, l0 int64, l1 int64, l2 int64, l3 int32, l4 int64) int32
	Ucol_getVersion_0(m *Module, l0 int64, l1 int64)
	U_versionToString_0(m *Module, l0 int64, l1 int64)
	U_errorName_0(m *Module, l0 int32) int64
	Nanosleep(m *Module, l0 int64, l1 int64) int32
	XmlXPathFreeObject(m *Module, l0 int64)
	XmlXPathFreeCompExpr(m *Module, l0 int64)
	XmlXPathFreeContext(m *Module, l0 int64)
	XmlFreeDoc(m *Module, l0 int64)
	XmlFreeParserCtxt(m *Module, l0 int64)
	XmlNewParserCtxt(m *Module) int64
	XmlCtxtReadMemory(m *Module, l0 int64, l1 int64, l2 int32, l3 int64, l4 int64, l5 int32) int64
	XmlXPathNewContext(m *Module, l0 int64) int64
	XmlXPathRegisterNs(m *Module, l0 int64, l1 int64, l2 int64) int32
	XmlXPathCtxtCompile(m *Module, l0 int64, l1 int64) int64
	XmlXPathCompiledEval(m *Module, l0 int64, l1 int64) int64
	XmlInitParser(m *Module)
	XmlSetExternalEntityLoader(m *Module, l0 int64)
	XmlSetStructuredErrorFunc(m *Module, l0 int64, l1 int64)
	XmlBufferContent(m *Module, l0 int64) int64
	XmlBufferLength(m *Module, l0 int64) int32
	XmlBufferCreate(m *Module) int64
	XmlNewTextWriterMemory(m *Module, l0 int64, l1 int32) int64
	XmlTextWriterStartElement(m *Module, l0 int64, l1 int64) int32
	XmlTextWriterWriteAttribute(m *Module, l0 int64, l1 int64, l2 int64) int32
	XmlTextWriterWriteRaw(m *Module, l0 int64, l1 int64) int32
	XmlTextWriterEndElement(m *Module, l0 int64) int32
	XmlFreeTextWriter(m *Module, l0 int64)
	XmlBufferFree(m *Module, l0 int64)
	XmlTextWriterWriteBinHex(m *Module, l0 int64, l1 int64, l2 int32, l3 int32) int32
	XmlTextWriterWriteBase64(m *Module, l0 int64, l1 int64, l2 int32, l3 int32) int32
	XmlStrlen(m *Module, l0 int64) int32
	XmlNewDoc(m *Module, l0 int64) int64
	XmlStrdup(m *Module, l0 int64) int64
	XmlKeepBlanksDefault(m *Module, l0 int32) int32
	XmlParseBalancedChunkMemory(m *Module, l0 int64, l1 int64, l2 int64, l3 int32, l4 int64, l5 int64) int32
	XmlCtxtReadDoc(m *Module, l0 int64, l1 int64, l2 int64, l3 int64, l4 int32) int64
	XmlEncodeSpecialChars(m *Module, l0 int64, l1 int64) int64
	XmlXPathCastNodeToString(m *Module, l0 int64) int64
	XmlCopyNode(m *Module, l0 int64, l1 int32) int64
	XmlNodeDump(m *Module, l0 int64, l1 int64, l2 int64, l3 int32, l4 int32) int32
	XmlFreeNode(m *Module, l0 int64)
	XmlSaveToBuffer(m *Module, l0 int64, l1 int64, l2 int32) int64
	XmlNewNode(m *Module, l0 int64, l1 int64) int64
	XmlDocSetRootElement(m *Module, l0 int64, l1 int64) int64
	XmlAddChildList(m *Module, l0 int64, l1 int64) int64
	XmlNewDocText(m *Module, l0 int64, l1 int64) int64
	XmlSaveClose(m *Module, l0 int64) int32
	XmlSaveTree(m *Module, l0 int64, l1 int64) int64
	XmlSaveDoc(m *Module, l0 int64, l1 int64) int64
	XmlXPathCastBooleanToString(m *Module, l0 int32) int64
	XmlXPathCastBooleanToNumber(m *Module, l0 int32) float64
	XmlXPathCastNumberToString(m *Module, l0 float64) int64
	XmlXPathCastNodeSetToString(m *Module, l0 int64) int64
}
type PgvfsImports interface {
	Host_now_ns(m *Module) int64
	Host_close(m *Module, l0 int32) int64
	Host_fsync(m *Module, l0 int32) int64
	Host_stdout(m *Module, l0 int64, l1 int64) int64
	Host_stdin(m *Module, l0 int64, l1 int64) int64
	Host_stderr(m *Module, l0 int64, l1 int64) int64
	Host_opendir(m *Module, l0 int64, l1 int64) int64
	Host_readdir(m *Module, l0 int32, l1 int64, l2 int64) int64
	Host_closedir(m *Module, l0 int32) int64
	Host_read(m *Module, l0 int32, l1 int64, l2 int64) int64
	Host_fstat(m *Module, l0 int32, l1 int64) int64
	Host_lseek(m *Module, l0 int32, l1 int64, l2 int32) int64
	Host_write(m *Module, l0 int32, l1 int64, l2 int64) int64
	Host_open(m *Module, l0 int64, l1 int64, l2 int32, l3 int32) int64
	Host_stat(m *Module, l0 int64, l1 int64, l2 int32, l3 int64) int64
	Host_mkdir(m *Module, l0 int64, l1 int64, l2 int32) int64
	Host_pread(m *Module, l0 int32, l1 int64, l2 int64, l3 int64) int64
	Host_rmdir(m *Module, l0 int64, l1 int64) int64
	Host_access(m *Module, l0 int64, l1 int64, l2 int32) int64
	Host_pwrite(m *Module, l0 int32, l1 int64, l2 int64, l3 int64) int64
	Host_rename(m *Module, l0 int64, l1 int64, l2 int64, l3 int64) int64
	Host_unlink(m *Module, l0 int64, l1 int64) int64
	Host_readlink(m *Module, l0 int64, l1 int64, l2 int64, l3 int64) int64
	Host_ftruncate(m *Module, l0 int32, l1 int64) int64
	Host_argc(m *Module) int64
	Host_argv(m *Module, l0 int32, l1 int64, l2 int64) int64
	Host_proc_exit(m *Module, l0 int32)
}
type Module struct {
	Memory []byte
	MaxMem uint64
	M      unsafe.Pointer
	T0     []any
	G0     int64
	Env    EnvImports
	Pgvfs  PgvfsImports
	MemMu  sync.Mutex
}

func I32(x int32) int32 { return x }

func I64(x int64) int64 { return x }

// ui32 / ui64 reinterpret a signed integer as its unsigned bit
// equivalent at runtime. Used for the operands of wasm unsigned
// comparisons (i32.lt_u etc.) — emitting `uint32(int32(-N))` directly
// fails Go's compile-time constant rule because the negative typed
// constant isn't representable in uint32; routing through these
// function-call boundaries forces runtime conversion.
func Ui32(x int32) uint32 { return uint32(x) }

func Ui64(x int64) uint64 { return uint64(x) }

// b2i32 materialises a wasm comparison result — an i32 that is 0 or 1 — from
// the Go bool the comparison expression evaluates to.
//
// It exists as a named helper rather than an inline `func() int32 { ... }()`
// because the gcasm backend requires every direct call left in the compiled
// output to be either a package-local FnN or something the Go inliner removed.
// A func literal is normally inlined at its call site, but the inliner gives up
// once the ENCLOSING function grows past its budget — and a single wasm function
// can translate to tens of thousands of lines of Go, as an interpreter's
// bytecode dispatch loop does. The literal is then outlined into a real closure
// symbol (FnN.funcA.funcB), which reaches the assembler as a direct call gcasm
// cannot marshal. A named helper this small is always inlined, and if it ever
// were not, it would fail loudly at its own symbol rather than as a nested
// closure.
func B2i32(b bool) int32 {
	if b {
		return 1
	}
	return 0
}

func F32(x float32) float32 { runtime.KeepAlive(&x); return x }

func F64(x float64) float64 { runtime.KeepAlive(&x); return x }

//go:noinline
func Wasm_trap_div_zero() { panic("wasm: integer divide by zero") }

//go:noinline
func Wasm_trap_int_overflow() { panic("wasm: integer overflow") }

//go:noinline
func Wasm_trap_invalid_conv() { panic("wasm: invalid conversion to integer") }

//go:noinline
func Wasm_trap_unreachable() { panic("wasm: unreachable") }

//go:noinline
func Wasm_trap_memfill_oob() { panic("wasm: memory.fill out of bounds") }

//go:noinline
func Wasm_trap_memcopy_oob() { panic("wasm: memory.copy out of bounds") }

func I32_div_s(x, y int32) int32 {
	if y == -1 && x == math.MinInt32 {
		Wasm_trap_int_overflow()
	}
	if y == 0 {
		Wasm_trap_div_zero()
	}
	return x / y
}

func I64_div_s(x, y int64) int64 {
	if y == -1 && x == math.MinInt64 {
		Wasm_trap_int_overflow()
	}
	if y == 0 {
		Wasm_trap_div_zero()
	}
	return x / y
}

func I32_div_u(x, y uint32) uint32 {
	if y == 0 {
		Wasm_trap_div_zero()
	}
	return x / y
}

func I64_div_u(x, y uint64) uint64 {
	if y == 0 {
		Wasm_trap_div_zero()
	}
	return x / y
}

func I32_rem_s(x, y int32) int32 {
	if y == 0 {
		Wasm_trap_div_zero()
	}
	if y == -1 {
		return 0
	}
	return x % y
}

func I64_rem_s(x, y int64) int64 {
	if y == 0 {
		Wasm_trap_div_zero()
	}
	if y == -1 {
		return 0
	}
	return x % y
}

func I32_rem_u(x, y uint32) uint32 {
	if y == 0 {
		Wasm_trap_div_zero()
	}
	return x % y
}

func I64_rem_u(x, y uint64) uint64 {
	if y == 0 {
		Wasm_trap_div_zero()
	}
	return x % y
}

func I32_rotl(x, y int32) int32 { return int32(bits.RotateLeft32(uint32(x), int(y&31))) }

func I32_rotr(x, y int32) int32 { return int32(bits.RotateLeft32(uint32(x), -int(y&31))) }

func I64_rotl(x, y int64) int64 { return int64(bits.RotateLeft64(uint64(x), int(y&63))) }

func I64_rotr(x, y int64) int64 { return int64(bits.RotateLeft64(uint64(x), -int(y&63))) }

func F64_min(x, y float64) float64 {
	if x != x || y != y {
		return math.NaN()
	}
	if x < y {
		return x
	}
	if y < x {
		return y
	}
	if x == 0 {
		if math.Signbit(x) {
			return x
		}
		return y
	}
	return x
}

func F32_abs(x float32) float32 { return math.Float32frombits(math.Float32bits(x) &^ (1 << 31)) }

func F64_abs(x float64) float64 { return math.Float64frombits(math.Float64bits(x) &^ (1 << 63)) }

func F32_neg(x float32) float32 { return math.Float32frombits(math.Float32bits(x) ^ (1 << 31)) }

func F64_neg(x float64) float64 { return math.Float64frombits(math.Float64bits(x) ^ (1 << 63)) }

func F64_copysign(x, y float64) float64 { return math.Copysign(x, y) }

func F32_nearest(x float32) float32 { return float32(math.RoundToEven(float64(x))) }

func F64_nearest(x float64) float64 { return math.RoundToEven(x) }

func I64_trunc_f64_u(x float64) int64 {
	if x != x {
		Wasm_trap_invalid_conv()
	}
	if !(x > -1.0 && x < 18446744073709551616.0) {
		Wasm_trap_int_overflow()
	}
	return int64(uint64(x))
}

func I32_trunc_sat_f32_s(x float32) int32 {
	if x != x {
		return 0
	}
	if x <= -2147483648.0 {
		return math.MinInt32
	}
	if x >= 2147483648.0 {
		return math.MaxInt32
	}
	return int32(x)
}

func I32_trunc_sat_f32_u(x float32) int32 {
	if x != x || x <= 0 {
		return 0
	}
	if x >= 4294967296.0 {
		return -1
	}
	return int32(uint32(x))
}

func I32_trunc_sat_f64_s(x float64) int32 {
	if x != x {
		return 0
	}
	if x <= -2147483648.0 {
		return math.MinInt32
	}
	if x >= 2147483648.0 {
		return math.MaxInt32
	}
	return int32(x)
}

func I32_trunc_sat_f64_u(x float64) int32 {
	if x != x || x <= 0 {
		return 0
	}
	if x >= 4294967296.0 {
		return -1
	}
	return int32(uint32(x))
}

func I64_trunc_sat_f32_s(x float32) int64 {
	if x != x {
		return 0
	}
	if float64(x) <= -9223372036854775808.0 {
		return math.MinInt64
	}
	if float64(x) >= 9223372036854775808.0 {
		return math.MaxInt64
	}
	return int64(x)
}

func I64_trunc_sat_f64_s(x float64) int64 {
	if x != x {
		return 0
	}
	if x <= -9223372036854775808.0 {
		return math.MinInt64
	}
	if x >= 9223372036854775808.0 {
		return math.MaxInt64
	}
	return int64(x)
}

func I64_trunc_sat_f64_u(x float64) int64 {
	if x != x || x <= 0 {
		return 0
	}
	if x >= 18446744073709551616.0 {
		return -1
	}
	return int64(uint64(x))
}

// memoryGrow grows m.memory by n wasm pages (64 KiB each). Returns the
// previous page count, or -1 if the new size would exceed maxMem. n may be 0,
// which simply returns the current size.
//
// len(m.memory) must always equal the exact wasm memory size (memory.size
// and every bounds check depend on it), but the backing array is grown
// GEOMETRICALLY: a sequence of small memory.grow calls — which a C++ heap
// does constantly during start-up — would otherwise reallocate and recopy
// the whole linear memory on every page, i.e. O(n^2) total copying. Spare
// capacity makes the common grow a zero-copy reslice and amortizes the
// reallocations to O(n).
func MemoryGrow(m *Module, n int32) int32 {

	m.MemMu.Lock()
	defer m.MemMu.Unlock()
	prev := int32(len(m.Memory) >> 16)
	if n == 0 {
		return prev
	}
	if n < 0 {
		return -1
	}

	want := uint64(len(m.Memory)) + uint64(n)*65536
	if m.MaxMem != 0 && want > m.MaxMem {
		return -1
	}
	if want > 1<<32 {
		return -1
	}
	if want <= uint64(cap(m.Memory)) {

		m.Memory = m.Memory[:want]
		return prev
	}

	newCap := uint64(cap(m.Memory)) * 2
	if newCap < want {
		newCap = want
	}
	if m.MaxMem != 0 && newCap > m.MaxMem {
		newCap = m.MaxMem
	}
	if newCap > 1<<32 {
		newCap = 1 << 32
	}

	grown := make([]byte, want, newCap)
	copy(grown, m.Memory)
	m.Memory = grown

	m.M = unsafe.Pointer(unsafe.SliceData(m.Memory))
	return prev
}

// accessMemory runs f with the module's current linear memory while
// holding the same lock memoryGrow takes to mutate the memory slice
// header or relocate its backing array. It is the ONE safe way to
// touch linear memory from OUTSIDE the module's execution goroutine —
// e.g. a watchdog goroutine raising CPython's eval-breaker bit while
// an evaluation is running. For the duration of f the memory can
// neither be resliced nor relocated, so f's writes land in the array
// the guest observes; a grow that raced in just before blocks until f
// returns and then copies f's writes forward with the rest of the
// contents. Determinism notes for callers:
//
//   - f MUST NOT call back into the module or into memoryGrow — that
//     would self-deadlock.
//   - f should be short: a running guest blocks inside memory.grow
//     until f returns (ordinary guest loads/stores do not block).
//   - Bytes the guest reads or writes concurrently with f (that is
//     the point of an eval-breaker-style flag) are exchanged with
//     plain single-word accesses; keep such shared words
//     word-aligned and word-sized.
func AccessMemory(m *Module, f func(mem []byte)) { m.MemMu.Lock(); defer m.MemMu.Unlock(); f(m.Memory) }

func I32_div_u_s(x, y int32) int32 { return int32(I32_div_u(uint32(x), uint32(y))) }
func I32_rem_u_s(x, y int32) int32 { return int32(I32_rem_u(uint32(x), uint32(y))) }
func I64_div_u_s(x, y int64) int64 { return int64(I64_div_u(uint64(x), uint64(y))) }
func I64_rem_u_s(x, y int64) int64 { return int64(I64_rem_u(uint64(x), uint64(y))) }

func F32_add(x, y float32) float32 { return x + y }
func F32_sub(x, y float32) float32 { return x - y }
func F32_mul(x, y float32) float32 { return x * y }
func F32_div(x, y float32) float32 { return x / y }
func F64_add(x, y float64) float64 { return x + y }
func F64_sub(x, y float64) float64 { return x - y }
func F64_mul(x, y float64) float64 { return x * y }
func F64_div(x, y float64) float64 { return x / y }

func I32_eqz(x int32) int32 {
	if x == 0 {
		return 1
	}
	return 0
}

func I64_eqz(x int64) int32 {
	if x == 0 {
		return 1
	}
	return 0
}

func I32_clz(x int32) int32    { return int32(bits.LeadingZeros32(uint32(x))) }
func I32_ctz(x int32) int32    { return int32(bits.TrailingZeros32(uint32(x))) }
func I32_popcnt(x int32) int32 { return int32(bits.OnesCount32(uint32(x))) }

func I64_clz(x int64) int64    { return int64(bits.LeadingZeros64(uint64(x))) }
func I64_ctz(x int64) int64    { return int64(bits.TrailingZeros64(uint64(x))) }
func I64_popcnt(x int64) int64 { return int64(bits.OnesCount64(uint64(x))) }

func F64_ceil(x float64) float64 { return math.Ceil(x) }

func F64_floor(x float64) float64 { return math.Floor(x) }
func F32_trunc(x float32) float32 { return float32(math.Trunc(float64(x))) }

func F32_sqrt(x float32) float32 { return float32(math.Sqrt(float64(x))) }
func F64_sqrt(x float64) float64 { return math.Sqrt(x) }

func F32_eq(x, y float32) int32 {
	if x == y {
		return 1
	}
	return 0
}

func F32_ne(x, y float32) int32 {
	if x != y {
		return 1
	}
	return 0
}

func F32_lt(x, y float32) int32 {
	if x < y {
		return 1
	}
	return 0
}

func F32_gt(x, y float32) int32 {
	if x > y {
		return 1
	}
	return 0
}

func F32_le(x, y float32) int32 {
	if x <= y {
		return 1
	}
	return 0
}

func F32_ge(x, y float32) int32 {
	if x >= y {
		return 1
	}
	return 0
}

func F64_eq(x, y float64) int32 {
	if x == y {
		return 1
	}
	return 0
}

func F64_ne(x, y float64) int32 {
	if x != y {
		return 1
	}
	return 0
}

func F64_lt(x, y float64) int32 {
	if x < y {
		return 1
	}
	return 0
}

func F64_gt(x, y float64) int32 {
	if x > y {
		return 1
	}
	return 0
}

func F64_le(x, y float64) int32 {
	if x <= y {
		return 1
	}
	return 0
}

func F64_ge(x, y float64) int32 {
	if x >= y {
		return 1
	}
	return 0
}

func I32_wrap_i64(x int64) int32       { return int32(x) }
func I64_extend_i32_s(x int32) int64   { return int64(x) }
func I64_extend_i32_u(x int32) int64   { return int64(uint32(x)) }
func F32_demote_f64(x float64) float32 { return float32(x) }
func F64_promote_f32(x float32) float64 {
	if math.IsNaN(float64(x)) {
		return float64(x)
	}
	return float64(x)
}

func F32_convert_i32_s(x int32) float32 { return float32(x) }
func F32_convert_i32_u(x int32) float32 { return float32(uint32(x)) }
func F32_convert_i64_s(x int64) float32 { return float32(x) }
func F32_convert_i64_u(x int64) float32 { return float32(uint64(x)) }
func F64_convert_i32_s(x int32) float64 { return float64(x) }
func F64_convert_i32_u(x int32) float64 { return float64(uint32(x)) }
func F64_convert_i64_s(x int64) float64 { return float64(x) }
func F64_convert_i64_u(x int64) float64 { return float64(uint64(x)) }

func I32_reinterpret_f32(x float32) int32 { return int32(math.Float32bits(x)) }
func I64_reinterpret_f64(x float64) int64 { return int64(math.Float64bits(x)) }
func F32_reinterpret_i32(x int32) float32 { return math.Float32frombits(uint32(x)) }
func F64_reinterpret_i64(x int64) float64 { return math.Float64frombits(uint64(x)) }

func I32_extend8_s(x int32) int32  { return int32(int8(x)) }
func I32_extend16_s(x int32) int32 { return int32(int16(x)) }
func I64_extend8_s(x int64) int64  { return int64(int8(x)) }
func I64_extend16_s(x int64) int64 { return int64(int16(x)) }
func I64_extend32_s(x int64) int64 { return int64(int32(x)) }

func MemoryFill(m *Module, dst int32, val int32, n int32) {
	if n == 0 {
		return
	}

	end := uint64(uint32(dst)) + uint64(uint32(n))
	if end > uint64(len(m.Memory)) {
		Wasm_trap_memfill_oob()
	}

	b := m.Memory[uint32(dst):uint32(end)]
	v := byte(val)
	if v == 0 {
		for k := range b {
			b[k] = 0
		}
		return
	}

	b[0] = v
	for filled := 1; filled < len(b); filled *= 2 {
		copy(b[filled:], b[:filled])
	}
}

func MemoryCopy(m *Module, dst int32, src int32, n int32) {
	if n == 0 {
		return
	}

	srcEnd := uint64(uint32(src)) + uint64(uint32(n))
	dstEnd := uint64(uint32(dst)) + uint64(uint32(n))
	if srcEnd > uint64(len(m.Memory)) || dstEnd > uint64(len(m.Memory)) {
		Wasm_trap_memcopy_oob()
	}

	copy(m.Memory[uint32(dst):uint32(dstEnd)], m.Memory[uint32(src):uint32(srcEnd)])
}
