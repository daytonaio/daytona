#!/bin/bash
set -e

pip install -e .

# finqa_env requires openenv-core>=0.2.1 but git HEAD is 0.2.0 (version bump pending).
# Install without deps to bypass the constraint — openenv-core is already installed above.
pip install --no-deps "openenv-finqa-env @ git+https://github.com/meta-pytorch/OpenEnv.git#subdirectory=envs/finqa_env"

echo ""
echo "Done! Next steps:"
echo "  cp ../.env.example ../.env   # add your DAYTONA_API_KEY"
echo "  python ../build_snapshot.py"
echo "  python train.py --sandboxes 2 --iterations 1 --group-size 2  # smoke test"
