if RUBY_PLATFORM =~ /darwin/
  EXE = './dynatlas'
else
  EXE = 'dynatlas.exe'
end


file EXE => Dir.glob('*.go') do
  # sh "go generate"
  sh "go build"
end

task :build => EXE

task :t1 => :build do
  sh EXE, 'fuga.png', 'fuga.tsp'
end

task :sample => :build do
  mkdir_p 'sample_tsp'
  Dir.glob('sample/*.png') do |f|
    sh EXE, f, 'sample_tsp/' + File.basename(f) + ".tsp"
  end
end
