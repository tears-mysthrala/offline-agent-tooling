Import-Module Pester -ErrorAction SilentlyContinue

# For Pester runs the current working directory is the repo root in our runner.
# Use Get-Location for simplicity and reliability.
$script:repo = (Get-Location).ProviderPath

Describe 'FS and Search tools (integration)' {

    Context 'FS operations' {
        It 'writes, reads and appends a file' {
            # determine repo at runtime (works inside Pester It block)
            $localRepo = (Get-Location).ProviderPath
            $tmp = Join-Path $localRepo 'tmp_pester.txt'
            $fs = Join-Path $localRepo 'tools\ps\fs.ps1'

            # write
            $out = & $fs -op write -path $tmp -content 'pester'
            $obj = $out | ConvertFrom-Json
            $obj.ok | Should -BeTrue

            # read
            $out = & $fs -op read -path $tmp
            $obj = $out | ConvertFrom-Json
            $obj.ok | Should -BeTrue
            $obj.data.content | Should -Match 'pester'

            # append
            & $fs -op append -path $tmp -content 'pester-append' | Out-Null

            # read-lines
            $out = & $fs -op read-lines -path $tmp
            $obj = $out | ConvertFrom-Json
            $obj.ok | Should -BeTrue
            $obj.data.lines.Count | Should -BeGreaterThan 1

            # checksum
            $out = & $fs -op checksum -path $tmp
            $obj = $out | ConvertFrom-Json
            $obj.ok | Should -BeTrue
            $obj.data.sha256 | Should -Not -BeNullOrEmpty

            # cleanup
            Remove-Item -Path $tmp -ErrorAction SilentlyContinue
        }
    }

    Context 'Search operations' {
        It 'finds TODO in fixtures sample.txt' {
            $localRepo = (Get-Location).ProviderPath
            $search = Join-Path $localRepo 'tools\ps\search.ps1'
            $out = & $search -op grep -root $localRepo -pattern 'TODO' -include '*.txt' -recursive
            $obj = $out | ConvertFrom-Json
            $obj.ok | Should -BeTrue
            $matches = $obj.data.matches | Where-Object { $_.file -match 'sample.txt' }
            $matches | Should -Not -BeNullOrEmpty
        }
    }
}
