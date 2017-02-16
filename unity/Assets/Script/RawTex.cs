using UnityEngine;
using System;
using System.Collections.Generic;
using System.IO;

public partial class DynAtlas {
	
	public class RawTex
	{
		public enum PKMTextureFormat {
			PKMFormatETC1 = 0
		}

		public enum TSPTextureFormat {
			PVRTC_RGBA4 = 1,
			ETC = 2
		}

		const int BlockLen = 4; // ブロックの大きさのピクセル数
		const int BlockSize = 8; // ブロックのサイズ[byte]

		public TextureFormat Format{ get; private set; }
		public int Width{ get; private set; }
		public int Height{ get; private set; }
		public byte[] Data { get ; private set; }
		public Texture Texture{ get; private set; }
		public bool IsRGB { get { return IsRGBFormat (Format); } }

		public RawTex(Texture2D tex){
			Format = tex.format;
			Texture = tex;
			Width = tex.width;
			Height = tex.height;
		}

		public RawTex(TextureFormat format, int width, int height, byte[] data = null)
		{
			Format = format;
			Width = width;
			Height = height;
			if( IsRGBFormat(format) ){
				Texture = new RenderTexture(width, height, 0, RenderTextureFormat.ARGB32, RenderTextureReadWrite.Default);
			}else{
				if (data != null)
				{
					Data = data;
				}
				else
				{
					Data = new byte[Width * Height / 2];
				}
			}
		}

		public void CopyRect(int srcX, int srcY, int width, int height, RawTex dest, int destX, int destY){
			if (Format == TextureFormat.PVRTC_RGBA4) {
				CopyRectToPVR (srcX, srcY, width, height, dest, destX, destY);
			} else if (Format == TextureFormat.ETC_RGB4) {
				CopyRectToETC (srcX, srcY, width, height, dest, destX, destY);
			} else if (IsRGB) {
				CopyRectToRGBA (srcX, srcY, width, height, dest, destX, destY);
			} else {
				throw new Exception ("unknown texture format");
			}
		}

		public void CopyRectToPVR(int srcX, int srcY, int width, int height, RawTex dest, int destX, int destY)
		{
			int srcBX = srcX / BlockLen;
			int srcBY = srcY / BlockLen;
			int destBX = destX / BlockLen;
			int destBY = destY / BlockLen;
			int blockWidth = width / BlockLen;
			int blockHeight = height / BlockLen;
			int bitSize = Mathf.FloorToInt (Mathf.Log (2048, 2));

			for (int by = 0; by < blockHeight; by++)
			{
				for (int bx = 0; bx < blockWidth; bx++)
				{
					var srcIndex = blockIndex(srcBX + bx, srcBY + by);
					var destIndex = dest.blockIndexPVR(bitSize, destBX + bx, destBY + by) * BlockSize;
					System.Array.Copy(Data, srcIndex, dest.Data, destIndex, BlockSize);
				}
			}
		}

		int blockIndexPVR(int bitSize, int bx, int by)
		{
			var r = 0;
			for( var i = 0; i < bitSize; i++ ){
				r |= (((bx & 1) << 1) | (by & 1)) << (i*2);
				bx = bx >> 1;
				by = by >> 1;
			}
			return r;
		}

		public void CopyRectToETC(int srcX, int srcY, int width, int height, RawTex dest, int destX, int destY)
		{
			int srcBX = srcX / BlockLen;
			int srcBY = srcY / BlockLen;
			int destBX = destX / BlockLen;
			int destBY = destY / BlockLen;
			int blockWidth = width / BlockLen;
			int blockHeight = height / BlockLen;

			for (int by = 0; by < blockHeight; by++)
			{
				for (int bx = 0; bx < blockWidth; bx++)
				{
					var srcIndex = blockIndex(srcBX + bx, srcBY + by);
					var destIndex = dest.blockIndex(destBX + bx, destBY + by);
					System.Array.Copy(Data, srcIndex, dest.Data, destIndex, BlockSize);
				}
			}
		}

		int blockIndex(int blockX, int blockY)
		{
			return (blockY * (Width / BlockLen) + blockX) * BlockSize;
		}

		public void CopyRectToRGBA(int srcX, int srcY, int width, int height, RawTex dest, int destX, int destY)
		{
			Debug.LogFormat ("{0} {1} {2} {3} {4} {5}", srcX, srcY, width, height, destX, destY);
			var rt = (RenderTexture)dest.Texture;
			Debug.Log (rt.IsCreated());
			Graphics.SetRenderTarget ((RenderTexture)dest.Texture);
			Graphics.CopyTexture (Texture, 0, 0, srcX, srcY, width, height, dest.Texture, 0, 0, destX, destY);
		}

		public static FileType FileTypeOfFilename(string filename){
			var extname = Path.GetExtension (filename);
			switch (extname.ToLowerInvariant ()) {
			case ".tsp":
				return FileType.TSP;
			case ".pkm":
				return FileType.PKM;
			case ".png":
				return FileType.PNG;
			default:
				throw new Exception ("unkonwn file extension " + filename);
			}
		}

		static ushort ntol(ushort n)
		{
			return (ushort)((n << 8) | (n >> 8));
		}

		public static RawTex Load(FileType fileType, Stream s){
			switch (fileType) {
			case FileType.TSP:
				return RawTex.LoadTSP (s);
			case FileType.PKM:
				return RawTex.LoadPKM (s);
			case FileType.PNG:
				return RawTex.LoadPNG (s);
			default:
				throw new Exception ("invalid file type");
			}
		}

		// PKMファイルを読み込む
		public static RawTex LoadPKM(Stream s)
		{
			using (var r = new BinaryReader(s))
			{
				var magic = r.ReadBytes(4);
				if (magic [0] != 'P' || magic [1] != 'K' || magic [2] != 'M' || magic [3] != ' ') {
					throw new Exception ("invalid PKM file header");
				}

				var version = r.ReadBytes(2);
				if (version [0] != '1' || version [1] != '0') {
					throw new Exception ("invalid PKM file version");
				}

				var format = r.ReadBytes(2);
				if (format [0] != 0 || format [1] != 0) {
					throw new Exception ("invalid PKM file format");
				}

				var width = ntol(r.ReadUInt16());
				var height = ntol(r.ReadUInt16());
				/*var origWidth =*/ ntol(r.ReadUInt16());
				/*var origHeight =*/ ntol(r.ReadUInt16());
				var data = r.ReadBytes(width * height / 2);

				return new RawTex(TextureFormat.ETC_RGB4, width, height, data);
			}

		}

		// TSPファイルを読み込む
		public static RawTex LoadTSP(Stream s) 
		{
			using (var r = new BinaryReader(s))
			{
				var magic = r.ReadBytes(4);
				if (magic [0] != 'T' || magic [1] != 'S' || magic [2] != 'P' || magic [3] != ' ') {
					throw new Exception ("invalid TSP file header");
				}

				var tspFormat = (TSPTextureFormat)r.ReadByte();
				TextureFormat format;
				switch( tspFormat ){
				case TSPTextureFormat.ETC:
					format = TextureFormat.ETC_RGB4;
					break;
				case TSPTextureFormat.PVRTC_RGBA4:
					format = TextureFormat.PVRTC_RGBA4;
					break;
				default:
					throw new Exception ("invalid TSP format " + tspFormat);
				}

				var width = r.ReadUInt16();
				var height = r.ReadUInt16();
				var data = r.ReadBytes(width * height / 2);

				return new RawTex(format, width, height, data);
			}
		}

		// PNGファイルを読み込む
		public static RawTex LoadPNG(Stream s) 
		{
			using (var r = new BinaryReader (s)) {
				byte[] data = r.ReadBytes ((int)r.BaseStream.Length);

				int pos = 16; // 16バイトから開始

				int width = 0;
				for (int i = 0; i < 4; i++) {
					width = width * 256 + data [pos++];
				}

				int height = 0;
				for (int i = 0; i < 4; i++) {
					height = height * 256 + data [pos++];
				}

				Texture2D tex = new Texture2D (width, height, TextureFormat.ARGB32, false, true);
				tex.LoadImage (data);

				LastLoaded = tex;

				return new RawTex (tex);
			}
		}

		public static Texture2D LastLoaded;
	}

}
