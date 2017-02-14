using UnityEngine;
using System.Collections;
using System.Collections.Generic;
using System.IO;

public class DynAtlas {

    int size_ = 2048;
    Texture2D tex_;
    Etc1Data data_;

    Dictionary<string, Sprite> sprites_ = new Dictionary<string,Sprite>();

    bool dirty_;

    public DynAtlas(int size)
    {
        size_ = size;
        tex_ = new Texture2D(size_, size_, TextureFormat.ETC_RGB4, false);
        data_ = new Etc1Data(size_, size_);
    }

    public Texture2D Texture { get { return tex_; } }

    static ushort ntol(ushort n)
    {
        return (ushort)((n << 8) | (n >> 8));
    }

    public void Load(string spriteName, Stream s)
    {
        using (var r = new BinaryReader(s))
        {
            Profiler.BeginSample("load");
            var magic = r.ReadBytes(4);
            var version = r.ReadBytes(2);
            var foramt = r.ReadBytes(2);
            var extendedWidth = ntol(r.ReadUInt16());
            var extendedHeight = ntol(r.ReadUInt16());
            var origWidth = ntol(r.ReadUInt16());
            var origHeight = ntol(r.ReadUInt16());
            var data = r.ReadBytes(extendedWidth * extendedHeight / 2);

            var rawtex = new Etc1Data(extendedWidth, extendedHeight, data);
            Profiler.EndSample();

            Profiler.BeginSample("copy");
            rawtex.CopyRect(0, 0, extendedWidth, extendedHeight, data_, 0, 0);
            Profiler.EndSample();

            sprites_[spriteName] = Sprite.Create(tex_, new Rect(0, 0, extendedWidth, extendedHeight), new Vector2(extendedWidth / 2f, extendedHeight / 2f));

            dirty_ = true;
        }

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


    public class Etc1Data
    {
        const int BlockLen = 4;
        const int BlockSize = 8;

        int width_;
        int height_;
        byte[] data_;

        public byte[] Data { get { return data_; } }

        public Etc1Data(int width, int height, byte[] data = null)
        {
            width_ = width;
            height_ = height;
            if (data != null)
            {
                data_ = data;
            }
            else
            {
                data_ = new byte[width_ * height_ / 2];
            }
        }

        public void CopyRect(int srcX, int srcY, int width, int height, Etc1Data dest, int destX, int destY )
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
                    //Debug.LogFormat("copy bx:{3} by:{4} src:{5:X} dest:{6:X} {0:X}=>{1:X}", srcIndex, destIndex, 0, bx, by, data_.Length, dest.data_.Length);
                    System.Array.Copy(data_, srcIndex, dest.data_, destIndex, BlockSize);
                }
            }
        }

        int blockIndex(int blockX, int blockY)
        {
            return (blockY * (width_ / BlockLen) + blockX) * BlockSize;
        }

    }

}
