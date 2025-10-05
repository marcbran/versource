local nl = import 'terraform-provider-null/main.libsonnet';
local tf = import 'terraform/main.libsonnet';

local splitNonEmpty(variables, key, default=null) =
  if std.get(variables, key, null) == null then default
  else [name for name in std.split(std.get(variables, key, ''), ',') if name != ''];

local module(variables) =
  local var = {
    keep: splitNonEmpty(variables, 'keep'),
    drop: splitNonEmpty(variables, 'drop'),
    names: splitNonEmpty(variables, 'names', ['a', 'b', 'c', 'd', 'e', 'f']),
    add: splitNonEmpty(variables, 'add'),
  };

  local resource(name) = nl.resource.resource(name, {
    triggers: {
      name: name,
    },
  });
  local resourcesMap = { [name]: resource(name) for name in var.names };
  local resources = std.objectValues(resourcesMap);

  local addResource(name) = {
    provider: 'null',
    resourceType: 'resource',
    name: name,
    attributes: {
      triggers: {
        name: name,
      },
    },
  };
  resources +
  [
    tf.Output('versource', {
      value: {
        keep: if var.keep == null then null else [std.get(resourcesMap, keep, null) for keep in var.keep],
        drop: if var.drop == null then null else [std.get(resourcesMap, drop, null) for drop in var.drop],
        add: if var.add == null then null else [addResource(name) for name in var.add],
      },
    }),
  ];

module
