local nl = import 'terraform-provider-null/main.libsonnet';
local tf = import 'terraform/main.libsonnet';

local module(var) =
  assert var.name != null : 'name required';

  local repo = nl.resource.resource('test', {
    triggers: {
      name: var.name,
    },
  });
  [
    tf.Output('id', {
      value: repo.id,
    }),
  ];

module
