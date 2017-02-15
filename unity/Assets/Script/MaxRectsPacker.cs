using System;
using System.Collections.Generic;
using UnityEngine;

public partial class DynAtlas {

	public class MaxRectsPacker : IPacker {
		MaxRectsBinPack pack_;

		public MaxRectsPacker(int width, int height)
		{
			pack_ = new MaxRectsBinPack(width, height, false);
		}

		public Vector2 Add(int w, int h)
		{
			var rect = pack_.Insert (w+4, h+4, MaxRectsBinPack.FreeRectChoiceHeuristic.RectBestLongSideFit);
			if (rect.width == 0 || rect.height == 0) {
				throw new Exception ("cannot pack");
			}
			return rect.min;
		}

		public void Delete()
		{
		}
	}
}
