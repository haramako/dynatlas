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
	}

	public interface IPacker {
		Vector2 Add(int w, int h);
	}

    int size_ = 2048;
    Texture2D tex_;
    //Etc1Data data_;
    RawTex data_;

    Dictionary<string, Sprite> sprites_ = new Dictionary<string,Sprite>();

    bool dirty_;
	IPacker packer_;

	static public TextureFormat DefaultTextureFormat()
	{
		return TextureFormat.PVRTC_RGBA4;
		//return TextureFormat.ETC_RGB4;
	}
		

    public DynAtlas(int size, IPacker packer = null)
    {
        size_ = size;
		tex_ = new Texture2D(size_, size_, DefaultTextureFormat(), false);
		tex_.wrapMode = TextureWrapMode.Clamp;
		data_ = new RawTex(DefaultTextureFormat(), size_, size_);
		//packer_ = new Packer (size_, size_);
		if (packer == null) {
			packer_ = new MaxRectsPacker (size_, size_);
		} else {
			packer_ = packer;
		}
    }

    public Texture2D Texture { get { return tex_; } }

    static ushort ntol(ushort n)
    {
        return (ushort)((n << 8) | (n >> 8));
    }

	public void Load(string filename, string spriteName = null ){
		if( spriteName == null ){
			spriteName = Path.GetFileNameWithoutExtension (filename);
		}
		using( var stream = File.OpenRead(filename) ){
			Load (spriteName, RawTex.FileTypeOfFilename(filename), stream);
		}
	}

    public void Load(string spriteName, FileType fileType, Stream s)
    {
        Profiler.BeginSample("load");
		RawTex rawtex = RawTex.Load(fileType, s);
        Profiler.EndSample();

		var pos = packer_.Add (rawtex.Width, rawtex.Height);

        Profiler.BeginSample("copy");
		rawtex.CopyRect(0, 0, rawtex.Width, rawtex.Height, data_, (int)pos.x, (int)pos.y);
        Profiler.EndSample();

		var sprite = Sprite.Create(
			tex_, 
			new Rect(pos.x, pos.y, rawtex.Width, rawtex.Height), 
			new Vector2(rawtex.Width / 2f, rawtex.Height / 2f), 
			100f, 0, SpriteMeshType.FullRect);
		
        sprites_[spriteName] = sprite;

        dirty_ = true;

    }

    public void ApplyChanges()
    {
        if( dirty_ )
        {
            dirty_ = false;

            Profiler.BeginSample("load tex");
            tex_.LoadRawTextureData(data_.Data);
            Profiler.EndSample();

            Profiler.BeginSample("apply");
            tex_.Apply();
            Profiler.EndSample();
        }
    }

	public Sprite FindSprite(string name)
	{
		Sprite found;
		if( sprites_.TryGetValue(name, out found))
		{
			return found;
		}
		else
		{
			return null;
		}
	}

	public IEnumerable<Sprite> GetSprites(){
		return sprites_.Values;
	}

}
