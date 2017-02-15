using UnityEngine;
using UnityEngine.UI;
using System.Collections;
using System.Collections.Generic;
using System.Linq;
using System.IO;
using System;

public class TestScene : MonoBehaviour {

    public DynAtlas atlas;
    public RawImage Image;
    public Image Img2;
	public Image Img3;
    public Sprite TemplateSprite;

	string slash(string path){
		return path.Replace ('/', Path.DirectorySeparatorChar);
	}

    // Use this for initialization
    IEnumerator Start () {

		yield return null;
		//yield return new WaitForSeconds (1.0f);

        atlas = new DynAtlas(4096);

		IEnumerable<string> dirs = new string[0];
		dirs = dirs.Concat( Directory.GetFiles(Path.Combine (Application.dataPath, slash ("../../sample_tsp"))));
		dirs = dirs.Concat( Directory.GetFiles(Path.Combine (Application.dataPath, slash ("../../sample2_tsp"))));

		int ii=0;
		try {
			foreach (var f in dirs ){
				if( f.EndsWith(".pvr.tsp") ){
					//Debug.Log(f);
					atlas.Load (f);
					ii++;
				}
			}
		}catch(Exception ex){
			Debug.LogException(ex);
			Debug.Log(ii);
		}

        atlas.ApplyChanges();

        Image.texture = atlas.Texture;
		Img2.sprite = atlas.FindSprite("fuga.pvr");
        Img3.sprite = atlas.FindSprite("piyo");

		while (true) {
			foreach (var sp in atlas.GetSprites()) {
				Img3.sprite = sp;
				Img3.SetNativeSize ();
				yield return new WaitForSeconds (0.2f);
			}
		}

	}
	
	// Update is called once per frame
	void Update () {
	
	}
}
