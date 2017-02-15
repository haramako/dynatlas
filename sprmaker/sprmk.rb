#!/usr/bin/env ruby
# Sprite Maker

require 'rake'

UNITY_EXE = "C:\\Program Files\\Unity\\Editor\\Unity.exe"

p pwd

rm_rf ['Assets/Work', 'Output']

mkdir_p ['Assets/Work', 'Output']

cp_r 'sample', 'Assets/Work'

begin
  sh(UNITY_EXE, '-batchmode', '-quit', '-logFile', 'build.log',
     '-projectPath', pwd,
     '-executeMethod', 'SprMaker.DoBatch')
rescue
  puts IO.binread('build.log').split(/\n/).last(10)
end

