#!/usr/bin/env ruby

require 'pp'
require 'rake'

PVRTOOL='bin/PVRTexToolCLI.exe'

f = "show.png"

if true

  sh "#{PVRTOOL} -f PVRTC1_4 -i #{f} -o fuga.pvr"
  

bin = File.open(f,'r:ascii-8bit')

pp bin.read(4).unpack('a4')
header = bin.read(48).unpack('i12')
flags, pixelformat0, pixelformat1, colorspace, channeltype, height, width, depth, surfaces, faces, mipmapcount, metadatasize = *header

pp bin.read(metadatasize)
pp header

pp bin.read(width*height/2).size

else
  # sh "etcpack.exe"
  sh "bin/etc1tool.exe #{f} -o fuga.pkm"
end
