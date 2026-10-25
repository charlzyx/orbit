fn main() {
 let mut cases=Vec::new();
 for (w,h) in [(1,1),(1,9),(11,1),(3,5),(17,13),(31,27)] {
  let im=image::RgbaImage::from_fn(w,h,|x,y|image::Rgba([((x*37+y*19)%256) as u8,((x*7+y*61+83)%256) as u8,((x*71+y*5+12)%256) as u8,((x*13+y*47+90)%256) as u8]));
  for half in 1..=12 {let sigma=half as f32/2.;let out=image::imageops::fast_blur(&im,sigma);cases.push(serde_json::json!({"width":w,"height":h,"sigma":sigma,"pixels":out.as_raw()}));}
 }
 println!("{}",serde_json::to_string(&cases).unwrap());
}
