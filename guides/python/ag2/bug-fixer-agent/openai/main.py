# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

import os

from autogen import ConversableAgent, LLMConfig  # pylint: disable=import-error
from autogen.coding import DaytonaCodeExecutor  # pylint: disable=import-error
from dotenv import load_dotenv  # pylint: disable=import-error

load_dotenv()

BUG_FIXER_SYSTEM_MESSAGE = """You are an expert bug fixer. You support Python, JavaScript, and TypeScript.
If asked to fix code in any other language, refuse and explain which languages are supported.

When given broken code:

1. Analyze the bug carefully and identify the root cause
2. Write the complete fixed code in a fenced code block using the correct language tag
3. Always include assertions or print statements at the end to verify the fix works
4. If your previous fix didn't work, analyze the error output and try a different approach
5. Once the code runs successfully, reply with just the word TERMINATE — never in the same message as a code block

Always wrap your code in fenced code blocks (```python, ```javascript, or ```typescript). Never explain without providing fixed code.
Never include TERMINATE in a message that contains a code block.
"""


def fix_bug(broken_code: str, error_description: str = "") -> None:
    """
    Fix broken code using AG2 agents with Daytona sandbox execution.

    The bug_fixer agent analyzes the code and proposes fixes, while the
    code_executor agent runs each attempt in an isolated Daytona sandbox.
    The loop continues until the code runs successfully or max attempts are reached.

    Args:
        broken_code: The broken code to fix.
        error_description: Optional description of the error or expected behavior.
    """
    llm_config = LLMConfig(
        {
            "model": "gpt-4o-mini",
            "api_key": os.environ["OPENAI_API_KEY"],
        }
    )

    with DaytonaCodeExecutor(timeout=60) as executor:
        bug_fixer = ConversableAgent(
            name="bug_fixer",
            system_message=BUG_FIXER_SYSTEM_MESSAGE,
            llm_config=llm_config,
            code_execution_config=False,
            is_termination_msg=lambda x: (
                "TERMINATE" in (x.get("content") or "") or not (x.get("content") or "").strip()
            ),
        )

        code_executor = ConversableAgent(
            name="code_executor",
            llm_config=False,
            code_execution_config={"executor": executor},
        )

        message = f"Fix this broken code:\n\n\n{broken_code}\n"
        if error_description:
            message += f"\n\nError: {error_description}"

        code_executor.run(
            recipient=bug_fixer,
            message=message,
            max_turns=8,
        ).process()


if __name__ == "__main__":
    # Example 1: Python — swapped operands in postfix expression evaluator
    broken_postfix = """\
def eval_postfix(expression):
    stack = []
    for token in expression.split():
        if token.lstrip('-').isdigit():
            stack.append(int(token))
        else:
            b = stack.pop()
            a = stack.pop()
            if token == '+':
                stack.append(a + b)
            elif token == '-':
                stack.append(b - a)
            elif token == '*':
                stack.append(a * b)
            elif token == '/':
                stack.append(b // a)
    return stack[0]

assert eval_postfix("3 4 +") == 7
assert eval_postfix("10 3 -") == 7, f"Got {eval_postfix('10 3 -')}"
assert eval_postfix("12 4 /") == 3, f"Got {eval_postfix('12 4 /')}"
assert eval_postfix("2 3 4 * +") == 14
print("All postfix tests passed!")
"""

    print("=" * 60)
    print("Example 1: Python — Postfix Expression Evaluator Bug")
    print("=" * 60)
    fix_bug(broken_postfix, "")

    # Example 2: JavaScript — wrong concatenation order in run-length encoder
    broken_js = """\
function encode(str) {
    if (!str) return '';
    let result = '';
    let count = 1;
    for (let i = 1; i < str.length; i++) {
        if (str[i] === str[i - 1]) {
            count++;
        } else {
            result += str[i - 1] + count;
            count = 1;
        }
    }
    result += str[str.length - 1] + count;
    return result;
}

console.assert(encode("aabbbcc") === "2a3b2c", `Expected "2a3b2c", got "${encode("aabbbcc")}"`);
console.assert(encode("abcd") === "1a1b1c1d", `Expected "1a1b1c1d", got "${encode("abcd")}"`);
console.log("All encoding tests passed!");
"""

    print("\n" + "=" * 60)
    print("Example 2: JavaScript — Run-Length Encoder Bug")
    print("=" * 60)
    fix_bug(broken_js, "")

    # Example 3: TypeScript — Math.min instead of Math.max in Kadane's algorithm
    broken_ts = """\
function maxSubarray(nums: number[]): number {
    let maxSum = nums[0];
    let currentSum = nums[0];
    for (let i = 1; i < nums.length; i++) {
        currentSum = Math.min(currentSum + nums[i], nums[i]);
        maxSum = Math.min(maxSum, currentSum);
    }
    return maxSum;
}

console.assert(maxSubarray([-2, 1, -3, 4, -1, 2, 1, -5, 4]) === 6,
    `Expected 6, got ${maxSubarray([-2, 1, -3, 4, -1, 2, 1, -5, 4])}`);
console.assert(maxSubarray([1]) === 1,
    `Expected 1, got ${maxSubarray([1])}`);
console.assert(maxSubarray([5, 4, -1, 7, 8]) === 23,
    `Expected 23, got ${maxSubarray([5, 4, -1, 7, 8])}`);
console.log("All max subarray tests passed!");
"""

    print("\n" + "=" * 60)
    print("Example 3: TypeScript — Max Subarray Bug")
    print("=" * 60)
    fix_bug(broken_ts, "")
