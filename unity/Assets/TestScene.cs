using UnityEngine;
using UnityEngine.UI;
using System.Collections;
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

        atlas = new DynAtlas(2048);

		string file;
		/*
		for (int i = 0; i < 30; i++) {
			if (Random.Range (0, 2) == 0) {
				file = Path.Combine (Application.dataPath, slash ("../../fuga.tsp"));
				atlas.Load (file);
			} else {
				file = Path.Combine (Application.dataPath, slash ("../../piyo.pkm"));
				atlas.Load (file);
			}
		}*/
		var dir = Path.Combine (Application.dataPath, slash ("../../sample_tsp"));
		try {
		foreach (var f in Directory.GetFiles(dir) ){
			Debug.Log(f);
			atlas.Load (f);
		}
		}catch(Exception ex){
		}

        atlas.ApplyChanges();

        Image.texture = atlas.Texture;
		Img2.sprite = atlas.FindSprite("fuga");
        Img3.sprite = atlas.FindSprite("piyo");
	}
	
	// Update is called once per frame
	void Update () {
	
	}
}
