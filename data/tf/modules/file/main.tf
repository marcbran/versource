resource "local_file" "test" {
  filename = "test.txt"
  content  = "testasdf"
}

output "test" {
  value = local_file.test.content
}
