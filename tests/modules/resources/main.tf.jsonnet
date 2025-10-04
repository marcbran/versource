local nl = import 'terraform-provider-null/main.libsonnet';
local tf = import 'terraform/main.libsonnet';

local module(variables) =
  local var = {
    keep: std.split(std.get(variables, 'keep', ''), ','),
    drop: std.split(std.get(variables, 'drop', ''), ','),
    names: std.split(std.get(variables, 'names', 'a,b,c,d,e,f'), ','),
  };

  local resource(name) = nl.resource.resource(name, {
    triggers: {
      name: name,
    },
  });
  local resourcesMap = { [name]: resource(name) for name in var.names };
  local resources = std.objectValues(resourcesMap);
  resources +
  [
    tf.Output('versource', {
      value: {
        keep: std.prune([std.get(resourcesMap, keep, null) for keep in var.keep]),
        drop: std.prune([std.get(resourcesMap, drop, null) for drop in var.drop]),
      },
    }),
  ];

module
