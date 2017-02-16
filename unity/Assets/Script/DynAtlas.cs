using UnityEngine;
using System;
using System.Collections;
using System.Collections.Generic;
using System.IO;
#if UNITY_5_5_OR_NEWER
using UnityEngine.Profiling;
#endif

public partial class DynAtlas {
	public enum FileType
	{
		TSP,
		PKM,
		PNG,
	}

	public interface IPacker {
		Vector2 Add(int w, int h);
	}

	public class DynSprite {
		public Atlas Atlas;
		public Sprite Sprite;
	}

	public int MaxSize { get; private set; }

	Dictionary<TextureFormat,List<Atlas>> atlases_ = new Dictionary<TextureFormat,List<Atlas>>();
	Dictionary<string, DynSprite> sprites_ = new Dictionary<string,DynSprite>();

	public DynAtlas(int maxSize = 0){
		MaxSize = maxSize;
	}

	public void ApplyChanges()
	{
		foreach (var list in atlases_.Values) {
			foreach (var atlas in list) {
				atlas.ApplyChanges ();
			}
		}
	}

	public Sprite Load(string filename, string spriteName = null ){
		if( spriteName == null ){
			spriteName = Path.GetFileNameWithoutExtension (filename);
		}
		using( var stream = File.OpenRead(filename) ){
			return Load (spriteName, RawTex.FileTypeOfFilename(filename), stream);
		}
	}

	public List<Atlas> GetAtlasesByFormat(TextureFormat format){
		List<Atlas> list;
		if (atlases_.TryGetValue (format, out list)) {
			return list;
		} else {
			list = new List<Atlas> ();
			atlases_ [format] = list;
			return list;
		}
	}

	public Sprite Load(string spriteName, FileType fileType, Stream s)
	{
		DynSprite result;

		Profiler.BeginSample("load");
		RawTex rawtex = RawTex.Load(fileType, s);
		Profiler.EndSample();

		var list = GetAtlasesByFormat (rawtex.Format);
		foreach (var atlas in list ) {
			result = addToAtlas (atlas, spriteName, rawtex);
			if (result != null ) {
				return result.Sprite;
			}
		}

		// 見つからなかった
		var newAtlas = new Atlas(rawtex.Format, MaxSize, new MaxRectsPacker(MaxSize, MaxSize));
		result = addToAtlas(newAtlas, spriteName, rawtex);

		list.Add (newAtlas);

		return result.Sprite;
	}

	DynSprite addToAtlas(Atlas atlas, string spriteName, RawTex rawtex){
		
		var pos = atlas.Add (rawtex);
		if (pos.x < 0) {
			return null;
		}

		var sprite = Sprite.Create(
			atlas.Texture, 
			new Rect(pos.x, pos.y, rawtex.Width, rawtex.Height), 
			new Vector2(rawtex.Width / 2f, rawtex.Height / 2f), 
			100f, 0, SpriteMeshType.FullRect);

		var dynSprite = new DynSprite {
			Atlas = atlas,
			Sprite = sprite,
		};

		sprites_ [spriteName] = dynSprite;

		return dynSprite;
	}

	public void ReserveTex(TextureFormat format){
		var list = GetAtlasesByFormat (format);
		var newAtlas = new Atlas(format, MaxSize, new MaxRectsPacker(MaxSize, MaxSize));
		list.Add (newAtlas);
	}


	public Sprite FindSprite(string name)
	{
		DynSprite found;
		if( sprites_.TryGetValue(name, out found))
		{
			return found.Sprite;
		}
		else
		{
			return null;
		}
	}

	public Dictionary<TextureFormat,List<Atlas>> GetAtlases(){
		return atlases_;
	}

	public IEnumerable<DynSprite> GetDynSprites(){
		return sprites_.Values;
	}


	static bool IsRGBFormat(TextureFormat format){
		return (format == TextureFormat.RGBA32 || format == TextureFormat.ARGB32 || format == TextureFormat.RGB24);
	}

}
