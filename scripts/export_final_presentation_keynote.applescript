on run argv
  if (count of argv) is less than 3 then
    error "Usage: osascript export_final_presentation_keynote.applescript input.pptx output.key output.pdf"
  end if

  set inputPath to item 1 of argv
  set keyPath to item 2 of argv
  set pdfPath to item 3 of argv

  set inputFile to POSIX file inputPath
  set keyFile to POSIX file keyPath
  set pdfFile to POSIX file pdfPath

  tell application "Keynote"
    activate
    set deck to open inputFile
    delay 2
    save deck in keyFile
    export deck to pdfFile as PDF
    close deck saving no
  end tell
end run
