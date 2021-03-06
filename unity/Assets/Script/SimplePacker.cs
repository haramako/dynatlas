﻿using System;
using System.Collections.Generic;
using UnityEngine;

public partial class DynAtlas {
	
	public class SimplePacker : IPacker {
		int width_;
		int height_;

		//int lowX;
		int lowY;
		int highX;
		int highY;

		public SimplePacker(int width, int height)
		{
			width_ = width;
			height_ = height;
			lowY = 4;
			highX = 4;
		}

		public Vector2 Add(int w, int h)
		{
			if (highX + w > width_) {
				highX = 4;
				lowY = highY;
			}

			if (lowY + h > height_) {
				return new Vector2 (-1, -1);
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
