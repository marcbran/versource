local Local = import 'terraform-provider-local/main.libsonnet';
local tf = import 'terraform/main.libsonnet';

local module(variables) =
  [
    Local.resource.file('failing_file', {
      filename: '/proc/readonly/filesystem/test.txt',
      content: 'This content will cause apply to fail',
    }),
    tf.Output('file_content', {
      value: '${local_file.failing_file.content}',
    }),
  ];

module
