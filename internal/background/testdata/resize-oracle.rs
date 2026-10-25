fn main() {
 let mut cases=Vec::new();
 for (w,h) in [(2001,17),(17,2001),(2001,1),(1,2001)] {
  let im=image::RgbaImage::from_fn(w,h,|x,y|image::Rgba([((x*37+y*19)%256) as u8,((x*7+y*61+83)%256) as u8,((x*71+y*5+12)%256) as u8,((x*13+y*47+90)%256) as u8]));
  let out=image::DynamicImage::ImageRgba8(im).resize(1600,1600,image::imageops::FilterType::Triangle).into_rgba8();cases.push(serde_json::json!({"width":w,"height":h,"out_width":out.width(),"out_height":out.height(),"pixels":out.as_raw()}));
 }
 println!("{}",serde_json::to_string(&cases).unwrap());
}
