resource "local_file" "test" {
  filename = "test.txt"
  content  = "test"
}

resource "local_file" "test2" {
  filename = "test2.txt"
  content  = "testasdfasdf!"
}

output "test" {
  value = local_file.test.content
}
