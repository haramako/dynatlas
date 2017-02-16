using UnityEngine;
using System;
using System.Collections;
using System.Collections.Generic;
using System.IO;
#if UNITY_5_5_OR_NEWER
using UnityEngine.Profiling;
#endif

public partial class DynAtlas {
	public class Atlas {
		int size_;
		Texture2D tex_;
		RawTex data_;

		public Texture2D Texture { get { return tex_; } }

		bool dirty_;
		IPacker packer_;

		public Atlas(TextureFormat format, int size = 0, IPacker packer = null)
		{
			size_ = size;
			tex_ = new Texture2D(size_, size_, format, false);
			tex_.wrapMode = TextureWrapMode.Clamp;
			data_ = new RawTex(format, size_, size_);
			//packer_ = new Packer (size_, size_);
			if (packer == null) {
				packer_ = new MaxRectsPacker (size_, size_);
			} else {
				packer_ = packer;
			}
		}

		/// <summary>
		/// アトラスにテクスチャを追加する
		/// </summary>
		/// <param name="rawtex">Rawtex.</param>
		public Vector2 Add(RawTex rawtex){
			var pos = packer_.Add (rawtex.Width, rawtex.Height);
			if (pos.x < 0) {
				return pos;
			}

			Profiler.BeginSample("copy");
			rawtex.CopyRect(0, 0, rawtex.Width, rawtex.Height, data_, (int)pos.x, (int)pos.y);
			Profiler.EndSample();

			dirty_ = true;

			return pos;
		}

		public void ApplyChanges()
		{
			if( dirty_ )
			{
				dirty_ = false;

				if (!IsRGBFormat (tex_.format)) {
					Profiler.BeginSample ("load tex");
					tex_.LoadRawTextureData (data_.Data);
					Profiler.EndSample ();
				}

				Profiler.BeginSample("apply");
				tex_.Apply();
				Profiler.EndSample();
			}
		}
	}
}
