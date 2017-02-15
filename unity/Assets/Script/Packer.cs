using System;
using System.Collections.Generic;
using UnityEngine;

public partial class DynAtlas {
	
	public class Packer {
		int width_;
		int height_;

		//int lowX;
		int lowY;
		int highX;
		int highY;

		public Packer(int width, int height)
		{
			width_ = width;
			height_ = height;
		}

		public Vector2 Add(int w, int h)
		{
			if (highX + w > width_) {
				highX = 0;
				lowY = highY;
			}

			if (lowY + h > height_) {
				throw new Exception ("can't insert");
			}

			var result = new Vector2 (highX, lowY);

			if (lowY + h > highY) {
				highY = lowY + h;
			}
			highX = highX + w;

			return result;
		}

		public void Delete()
		{
		}
	}
}
