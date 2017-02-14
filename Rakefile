
EXE = 'dynatlas.exe'

file EXE => Dir.glob('*.go') do
  # sh "go generate"
  sh "go build"
end

task :build => EXE

task :t1 => :build do
  sh EXE, 'fuga.png', 'fuga.tsp'
end
