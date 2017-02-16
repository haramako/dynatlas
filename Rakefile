# coding: utf-8
if RUBY_PLATFORM =~ /darwin/
  EXE_EXT = ''
  EXE = './bin/png2tsp'
else
  EXE_EXT = '.exe'
  EXE = './bin/png2tsp.exe'
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
end

task :sample => :build do
  mkdir_p 'sample_tsp'
  Dir.glob('sample/*.png') do |f|
    sh EXE, f, 'sample_tsp/' + File.basename(f) + ".etc.tsp"
    sh EXE, '-f', 'pvrtc', f, 'sample_tsp/' + File.basename(f) + ".pvr.tsp"
  end
end

task :sample2 => :build do
  mkdir_p 'sample2_tsp'
  Dir.glob('sample2/*.png') do |f|
    sh EXE, f, 'sample2_tsp/' + File.basename(f) + ".etc.tsp"
    sh EXE, '-f', 'PVRTC', f, 'sample2_tsp/' + File.basename(f) + ".pvr.tsp"
  end
end
