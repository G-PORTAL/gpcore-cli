# gon.hcl
#
# The path follows a pattern
# ./dist/BUILD-ID_TARGET/BINARY-NAME
source = ["./dist/gpcore-macos_darwin_amd64/gpcore","./dist/gpcore-macos_darwin_arm64/gpcore"]
bundle_id = "io.gpcore.cli"

apple_id {
  username = "@env:APPLE_EMAIL"
  password = "@env:APPLE_PASSWORD"
}

sign {
  application_identity = "Developer ID Application: Ociris GmbH"
}
