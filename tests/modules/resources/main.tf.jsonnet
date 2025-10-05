local nl = import 'terraform-provider-null/main.libsonnet';
local tf = import 'terraform/main.libsonnet';

local module(variables) =
  local var = {
    keep: std.split(std.get(variables, 'keep', ''), ','),
    drop: std.split(std.get(variables, 'drop', ''), ','),
    names: std.split(std.get(variables, 'names', 'a,b,c,d,e,f'), ','),
    add: std.split(std.get(variables, 'add', ''), ','),
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
  local addResources = [addResource(name) for name in var.add if name != ''];

  resources +
  [
    tf.Output('versource', {
      value: {
        keep: if std.length(var.keep) > 0 then std.prune([std.get(resourcesMap, keep, null) for keep in var.keep]) else null,
        drop: if std.length(var.drop) > 0 then std.prune([std.get(resourcesMap, drop, null) for drop in var.drop]) else null,
        add: if std.length(addResources) > 0 then addResources else null,
      },
    }),
  ];

module
