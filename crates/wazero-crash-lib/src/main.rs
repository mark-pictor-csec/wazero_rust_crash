extern crate alloc;

use std::alloc::{alloc, dealloc, Layout};

use aho_corasick::AhoCorasickBuilder;
use regex::Regex;

fn main() {
    let hw = "Hello, world! vectors";
    if !regex(hw) {
        println!("not re match");
    }
    if !ac(hw) {
        println!("not ac match")
    }
}

/*
in a wasm debug build, calling this function from wazero v1.7.{0,1} causes a crash
does not crash in release build

2024/05/08 13:21:27 call regex(): wasm error: unreachable
wasm stack trace:
        wazero_crash_lib-fcc161339c871004.wasm.abort()
        wazero_crash_lib-fcc161339c871004.wasm._ZN3std3sys3pal4wasi7helpers14abort_internal17h05344e3339eea616E()
                0x353106: /rustc/9b00956e56009bab2aa15d7bff10916599e3d6d6/library/std/src/sys/pal/wasi/helpers.rs:107:14
        wazero_crash_lib-fcc161339c871004.wasm._ZN3std9panicking20rust_panic_with_hook17hd3fb69bc0aea298aE(i32,i32,i32,i32,i32,i32)
                0x35552a: /rustc/9b00956e56009bab2aa15d7bff10916599e3d6d6/library/std/src/panicking.rs:798:9
        wazero_crash_lib-fcc161339c871004.wasm._ZN3std9panicking19begin_panic_handler28_$u7b$$u7b$closure$u7d$$u7d$17h4d99b90b43f79472E(i32)
                0x35474b: /rustc/9b00956e56009bab2aa15d7bff10916599e3d6d6/library/std/src/panicking.rs:649:13
        wazero_crash_lib-fcc161339c871004.wasm._ZN3std10sys_common9backtrace26__rust_end_short_backtrace17h5691573a73161cb1E(i32)
                0x354677: /rustc/9b00956e56009bab2aa15d7bff10916599e3d6d6/library/std/src/sys_common/backtrace.rs:171:18
        wazero_crash_lib-fcc161339c871004.wasm.rust_begin_unwind(i32)
                0x354e77: /rustc/9b00956e56009bab2aa15d7bff10916599e3d6d6/library/std/src/panicking.rs:645:5
        wazero_crash_lib-fcc161339c871004.wasm._ZN4core9panicking18panic_nounwind_fmt17hf5a5001f1ed7aacfE(i32,i32,i32)
                0x35c653: /rustc/9b00956e56009bab2aa15d7bff10916599e3d6d6/library/core/src/panicking.rs:110:18 (inlined)
                          /rustc/9b00956e56009bab2aa15d7bff10916599e3d6d6/library/core/src/panicking.rs:123:9
        wazero_crash_lib-fcc161339c871004.wasm._ZN4core9panicking14panic_nounwind17h3097dfdd0babd915E(i32,i32)
                0x35c6ab: /rustc/9b00956e56009bab2aa15d7bff10916599e3d6d6/library/core/src/panicking.rs:156:5
        wazero_crash_lib-fcc161339c871004.wasm._ZN4core10intrinsics19copy_nonoverlapping18precondition_check17h43ba69f190921943E(i32,i32,i32,i32,i32)
                0x347e6c: /rustc/9b00956e56009bab2aa15d7bff10916599e3d6d6/library/core/src/intrinsics.rs:2799:21
        wazero_crash_lib-fcc161339c871004.wasm._ZN4core3ptr14read_unaligned17h2e068735b35918d5E(i32) i32
                0x3487d7: /rustc/9b00956e56009bab2aa15d7bff10916599e3d6d6/library/core/src/intrinsics.rs:2969:5 (inlined)
                          /rustc/9b00956e56009bab2aa15d7bff10916599e3d6d6/library/core/src/ptr/mod.rs:1382:9
        wazero_crash_lib-fcc161339c871004.wasm._ZN4core3ptr9const_ptr33_$LT$impl$u20$$BP$const$u20$T$GT$14read_unaligned17hc093216b7e2fcf99E(i32) i32
                0x3489c1: /rustc/9b00956e56009bab2aa15d7bff10916599e3d6d6/library/core/src/ptr/const_ptr.rs:1296:18
        wazero_crash_lib-fcc161339c871004.wasm._ZN6memchr4arch3all6memchr3One8find_raw17h892cb0cf642a07baE(i32,i32,i32,i32)
                0x349c13: /path/to/.cargo/registry/src/index.crates.io-6f17d22bba15001f/memchr-2.7.2/src/arch/all/memchr.rs:143:21
        wazero_crash_lib-fcc161339c871004.wasm._ZN6memchr6memchr10memchr_raw17h8b9426aaa09e7975E(i32,i32,i32,i32)
                0x34ea05: /path/to/.cargo/registry/src/index.crates.io-6f17d22bba15001f/memchr-2.7.2/src/memchr.rs:531:9
        wazero_crash_lib-fcc161339c871004.wasm._ZN6memchr6memchr6memchr28_$u7b$$u7b$closure$u7d$$u7d$17hdbe3d4005e2d9c98E(i32,i32,i32,i32)
                0x34e94a: /path/to/.cargo/registry/src/index.crates.io-6f17d22bba15001f/memchr-2.7.2/src/memchr.rs:32:13
        wazero_crash_lib-fcc161339c871004.wasm._ZN6memchr6memchr6memchr17h9a8051a3e6827af1E(i32,i32,i32,i32)
                0x34ad6c: /path/to/.cargo/registry/src/index.crates.io-6f17d22bba15001f/memchr-2.7.2/src/arch/generic/memchr.rs:1134:17 (inlined)
                          /path/to/.cargo/registry/src/index.crates.io-6f17d22bba15001f/memchr-2.7.2/src/memchr.rs:31:9
        wazero_crash_lib-fcc161339c871004.wasm._ZN6memchr4arch3all10packedpair6Finder14find_prefilter17hec6b36d475802fd8E(i32,i32,i32,i32)
                0x34a9ca: /path/to/.cargo/registry/src/index.crates.io-6f17d22bba15001f/memchr-2.7.2/src/arch/all/packedpair/mod.rs:77:18
        wazero_crash_lib-fcc161339c871004.wasm._ZN6memchr6memmem8searcher23prefilter_kind_fallback17hc5b9b6f86b012be0E(i32,i32,i32,i32)
                0x350c78: /path/to/.cargo/registry/src/index.crates.io-6f17d22bba15001f/memchr-2.7.2/src/memmem/searcher.rs:789:5
        wazero_crash_lib-fcc161339c871004.wasm._ZN6memchr6memmem8searcher9Prefilter4find17hf3bb73a6146ed4baE(i32,i32,i32,i32)
                0x350bed: /path/to/.cargo/registry/src/index.crates.io-6f17d22bba15001f/memchr-2.7.2/src/memmem/searcher.rs:720:18
        wazero_crash_lib-fcc161339c871004.wasm._ZN6memchr6memmem8searcher3Pre4find17hbd6b997fa27732beE(i32,i32,i32,i32)
                0x34d720: /path/to/.cargo/registry/src/index.crates.io-6f17d22bba15001f/memchr-2.7.2/src/memmem/searcher.rs:971:22
        wazero_crash_lib-fcc161339c871004.wasm._ZN6memchr6memmem8searcher36searcher_kind_two_way_with_prefilter17hd9b46f49b15fc8f0E(i32,i32,i32,i32,i32,i32,i32)
                0x350007: /path/to/.cargo/registry/src/index.crates.io-6f17d22bba15001f/memchr-2.7.2/src/arch/all/twoway.rs:241:28 (inlined)
                          /path/to/.cargo/registry/src/index.crates.io-6f17d22bba15001f/memchr-2.7.2/src/arch/all/twoway.rs:162:17 (inlined)
                          /path/to/.cargo/registry/src/index.crates.io-6f17d22bba15001f/memchr-2.7.2/src/memmem/searcher.rs:351:9
        wazero_crash_lib-fcc161339c871004.wasm._ZN6memchr6memmem6Finder4find17h943d86fb8c506c05E(i32,i32,i32,i32)
                0x7f7ce: /path/to/.cargo/registry/src/index.crates.io-6f17d22bba15001f/memchr-2.7.2/src/memmem/searcher.rs:222:22 (inlined)
                         /path/to/.cargo/registry/src/index.crates.io-6f17d22bba15001f/memchr-2.7.2/src/memmem/mod.rs:427:9
        wazero_crash_lib-fcc161339c871004.wasm._ZN111_$LT$regex_automata..util..prefilter..memmem..Memmem$u20$as$u20$regex_automata..util..prefilter..PrefilterI$GT$4find17h167ea31f9222157dE(i32,i32,i32,i32,i32,i32)
                0x10de98: /path/to/.cargo/registry/src/index.crates.io-6f17d22bba15001f/regex-automata-0.4.6/src/util/prefilter/memmem.rs:43:13
        wazero_crash_lib-fcc161339c871004.wasm._ZN105_$LT$regex_automata..meta..strategy..Pre$LT$P$GT$$u20$as$u20$regex_automata..meta..strategy..Strategy$GT$11search_half17h5250af96968c9905E(i32,i32,i32,i32)
                0x148a0d: /path/to/.cargo/registry/src/index.crates.io-6f17d22bba15001f/regex-automata-0.4.6/src/meta/strategy.rs:393:9 (inlined)
                          /path/to/.cargo/registry/src/index.crates.io-6f17d22bba15001f/regex-automata-0.4.6/src/meta/strategy.rs:404:9
        wazero_crash_lib-fcc161339c871004.wasm._ZN14regex_automata4meta5regex5Regex11search_half17h672a4b5017868d1fE(i32,i32,i32)
                0x93f5: /path/to/.cargo/registry/src/index.crates.io-6f17d22bba15001f/regex-automata-0.4.6/src/meta/regex.rs:980:22
        wazero_crash_lib-fcc161339c871004.wasm._ZN5regex5regex6string5Regex11is_match_at17h43e8b217a45b9696E(i32,i32,i32,i32) i32
                0xed7b: /path/to/.cargo/registry/src/index.crates.io-6f17d22bba15001f/regex-1.10.4/src/regex/string.rs:1074:9
        wazero_crash_lib-fcc161339c871004.wasm._ZN5regex5regex6string5Regex8is_match17h01178aa758bc7529E(i32,i32,i32) i32
                0xedfe: /path/to/.cargo/registry/src/index.crates.io-6f17d22bba15001f/regex-1.10.4/src/regex/string.rs:205:9
        wazero_crash_lib-fcc161339c871004.wasm.regex(i32,i32) i32
                0x10459: /path/to/ahocorasick_crash/crates/wazero-crash-lib/src/main.rs:65:5
exit status 1
 */
#[cfg_attr(all(target_arch = "wasm32"), export_name = "regex")]
fn regex(s: &str) -> bool {
    let pattern = Regex::new(r#"orld"#).unwrap();
    pattern.is_match(s)
}

#[cfg_attr(all(target_arch = "wasm32"), export_name = "ac")]
fn ac(s: &str) -> bool {
    const SHORT_VECTORS: [&str; 2] = ["short", "vectors"];
    let ac = AhoCorasickBuilder::new()
        .ascii_case_insensitive(true)
        .build(SHORT_VECTORS)
        .unwrap();

    ac.is_match(s)
}

#[cfg_attr(all(target_arch = "wasm32"), export_name = "allocate")]
#[no_mangle]
pub unsafe extern "C" fn _allocate(size: u32) -> *mut u8 {
    allocate(size as usize)
}

unsafe fn allocate(size: usize) -> *mut u8 {
    let layout = Layout::from_size_align(size, std::mem::align_of::<u8>()).expect("Bad layout");
    alloc(layout)
}

#[cfg_attr(all(target_arch = "wasm32"), export_name = "deallocate")]
#[no_mangle]
pub unsafe extern "C" fn _deallocate(ptr: u32, size: u32) {
    let layout =
        Layout::from_size_align(size as usize, std::mem::align_of::<u8>()).expect("Bad layout");
    dealloc(ptr as *mut u8, layout)
}
