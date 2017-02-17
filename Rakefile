# coding: utf-8
require 'find'

UNITY = Dir.glob('/Applications/Unity5.5*/Unity.app').first 
MCS = UNITY + '/Contents/Mono/bin/smcs'

if RUBY_PLATFORM =~ /darwin/
  EXE_EXT = ''
  EXE = './bin/png2tsp'
else
  EXE_EXT = '.exe'
  EXE = './bin/png2tsp.exe'
end

def make_dll(dir, out, defines)
  cs_files = Find.find(dir).select{|f| File.extname(f) == '.cs' }
  sh MCS,
     "-r:#{UNITY}/Contents/Managed/UnityEngine.dll",
     "-r:#{UNITY}/Contents/Managed/UnityEditor.dll",
     "-r:#{UNITY}/Contents/UnityExtensions/Unity/GUISystem/UnityEngine.UI.dll",
     "-target:library",
     "-out:#{out}",
     *defines.map{|x| "-define:#{x}" },
     *cs_files
end

desc 'C#のDLLを作成する'
task :dll do
  make_dll('unity/Assets/Script', 'DynAtlas.dll', ["UNITY_5_5_OR_NEWER", "UNITY_ANDROID"])
end

file EXE => Dir.glob('png2tsp/*.go') do
  chdir 'png2tsp' do
    # sh "go generate"
    sh "go build"
  end
  cp 'png2tsp/png2tsp'+EXE_EXT, EXE
end

desc '実行ファイルを作成する'
task :build => EXE

task :test => :build do
  sh( EXE, '-h' ) rescue nil
  sh EXE, 'fuga.png', 'fuga.tsp'
  sh EXE, '-f', 'PVRTC', 'fuga.png', 'fuga.pvr.tsp'

  mkdir_p 'outdirtest'
  sh EXE, '-batch', '-outdir=outdirtest', '-postfix=-etc.tsp', '-f=ETC1', 'fuga.png', 'fugahalf.png'
  sh EXE, '-batch', '-outdir=outdirtest', '-postfix=-pvr.tsp', '-f=PVRTC', 'fuga.png', 'fugahalf.png'
end

task :sample => :build do
  mkdir_p 'sample_tsp'
  files = Dir.glob('sample/*.png')
  sh EXE, '-batch', '-f=ETC1', '-outdir=sample_tsp', '-postfix=.etc.tsp', *files
  sh EXE, '-batch', '-f=PVRTC', '-outdir=sample_tsp', '-postfix=.pvr.tsp', *files
end

task :sample2 => :build do
  mkdir_p 'sample2_tsp'
  files = Dir.glob('sample2/*.png')
  sh EXE, '-batch', '-f=ETC1', '-outdir=sample2_tsp', '-postfix=.etc.tsp', *files
  sh EXE, '-batch', '-f=PVRTC', '-outdir=sample2_tsp', '-postfix=.pvr.tsp', *files
end
