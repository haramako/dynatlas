using UnityEngine;
using UnityEditor;
using System.IO;

public class SprMaker {

    [MenuItem("Window/SprMaker/Build")]
    static void Open()
    {
        try
        {
            setAssetBundleName();

            buildAssetBundle();
        }
        finally
        {
            AssetDatabase.Refresh();
        }
    }

    public static void DoBatch()
    {
        try
        {
            setAssetBundleName();

            buildAssetBundle();
        }
        finally
        {
            AssetDatabase.Refresh();
        }
    }

    static void setAssetBundleName()
    {
        var importer = AssetImporter.GetAtPath("Assets/Work");
        importer.assetBundleName = "output.ab";
    }

    static void buildAssetBundle()
    {
        var opt = BuildAssetBundleOptions.IgnoreTypeTreeChanges;
        BuildPipeline.BuildAssetBundles("Output", opt, BuildTarget.Android);
    }
}
