default:
    @just --list

[no-cd]
jsonnet-test:
    #!/usr/bin/env bash
    exit_code=0
    if ! jsonnet ./tests.jsonnet > /dev/null; then
      echo 'Cannot execute jsonnet!'
      exit 1
    fi
    while IFS=$'\t' read -r name actual expected; do
      if [ "${actual}" != "${expected}" ]; then
        echo "${name} failed!"
        echo "${expected} was expected"
        echo "${actual} was the actual value"
        echo ""
        exit_code=1
      fi
    done < <(jsonnet ./tests.jsonnet | jq -r 'map([.name, (.actual | tostring), (.expected | tostring)]) | .[] | @tsv')

    if [ "${exit_code}" -eq 0 ]; then
      echo "All tests passed!"
    else
      echo "Some tests failed!"
    fi
    exit "${exit_code}"

[no-cd]
jsonnet-release branch path="" source=".":
    #!/usr/bin/env bash
    branch="{{branch}}"
    path="{{path}}"
    source="{{source}}"

    if [[ "${path}" == "" ]]; then
      path="${branch}"
    fi

    rm -rf release
    git clone git@github.com:marcbran/jsonnet.git release

    pushd release
    git checkout "${branch}" || git checkout -b "${branch}"
    git pull
    popd

    mkdir -p "release/${path}"
    cp "${source}/main.libsonnet" "release/${path}/main.libsonnet"

    pushd release
    git add -A
    git commit -m "release ${path}"
    git push --set-upstream origin "${branch}"
    popd

    rm -rf release
