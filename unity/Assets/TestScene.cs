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

		Debug.Log (SystemInfo.copyTextureSupport);

		yield return null;
		//yield return new WaitForSeconds (1.0f);

		atlas = new DynAtlas(1024);
		atlas.ReserveTex (TextureFormat.ARGB32);

		yield return new WaitForSeconds (0.5f);

		IEnumerable<string> dirs = new string[0];
		//dirs = dirs.Concat( Directory.GetFiles(Path.Combine (Application.dataPath, slash ("../../sample_tsp"))));
		//dirs = dirs.Concat( Directory.GetFiles(Path.Combine (Application.dataPath, slash ("../../sample2_tsp"))));
		dirs = dirs.Concat( Directory.GetFiles(Path.Combine (Application.dataPath, slash ("../../sample2"))));

		int ii=0;
		// {
			foreach (var f in dirs ){
				if( f.EndsWith(".png" /*".tsp" */ /*".pvr.tsp"*/) ){
					Debug.Log(f);
					atlas.Load (f);
					yield return new WaitForSeconds (0.2f);
					ii++;
					if( ii > 3 ) break;
				}
			}
		//}catch(Exception ex){
		//	Debug.LogException(ex);
		//	Debug.Log(ii);
		//}

        atlas.ApplyChanges();

		atlasList = atlas.GetAtlases ().Values.SelectMany (i => i).ToList ();
		OnImageClick ();

		Image.texture = DynAtlas.RawTex.LastLoaded;

		while (true) {
			foreach (var sp in atlas.GetDynSprites()) {
				Img3.sprite = sp.Sprite;
				Img3.SetNativeSize ();
				yield return new WaitForSeconds (0.2f);
			}
			yield return new WaitForSeconds (0.2f);
		}

	}

	List<DynAtlas.Atlas> atlasList;
	int i=0;

	public void OnImageClick(){
		Image.texture = atlasList [i].Texture;
		i++;
		if (i >= atlasList.Count) {
			i = 0;
		}
	}
	

}
