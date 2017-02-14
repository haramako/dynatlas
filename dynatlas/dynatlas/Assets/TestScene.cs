using UnityEngine;
using UnityEngine.UI;
using System.Collections;
using System.IO;

public class TestScene : MonoBehaviour {

    public DynAtlas atlas;
    public RawImage Image;
    public Image Img2;

	// Use this for initialization
	void Start () {
        atlas = new DynAtlas(2048);

        var file = new FileStream(Path.Combine(Application.dataPath, "..\\..\\..\\fuga.pkm"), FileMode.Open);
        atlas.Load("fuga", file);

        atlas.ApplyChanges();

        Image.texture = atlas.Texture;
        Img2.sprite = atlas.FindSprite("fuga");
	}
	
	// Update is called once per frame
	void Update () {
	
	}
}
