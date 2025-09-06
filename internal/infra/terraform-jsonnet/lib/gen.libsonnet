local tf = import 'terraform/main.libsonnet';

local terraformModuleDir(module, var) = tf.CfgDir(module(var));

terraformModuleDir
