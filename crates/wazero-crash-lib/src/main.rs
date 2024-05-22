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
