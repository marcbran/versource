local tf = import 'terraform/main.libsonnet';

local terraformModuleDir(module, var, statePath) =
  local evaluatedModule = module(var);
  local stack = if std.type(evaluatedModule) == 'array' then evaluatedModule else [evaluatedModule];
  local stateConfig = { terraform: { backend: { 'local': { path: statePath } } } };
  { 'main.tf.json': std.manifestJson(tf.Cfg(stack) + [stateConfig]) };

terraformModuleDir
