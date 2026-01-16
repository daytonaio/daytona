# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

"""Sandbox budget tracking for deeper-rlm."""

import threading
from dataclasses import dataclass


@dataclass
class BudgetStatus:
    """Current budget status."""

    max_sandboxes: int
    created: int
    active: int
    remaining: int


class SandboxBudget:
    """
    Thread-safe sandbox budget tracker.

    Tracks the total number of sandboxes that can be created across the entire
    RLM tree, regardless of depth level.
    """

    def __init__(self, max_sandboxes: int):
        """
        Initialize the budget tracker.

        Args:
            max_sandboxes: Maximum total sandboxes that can be created
        """
        self.max = max_sandboxes
        self.created = 0
        self.active = 0
        self._lock = threading.Lock()

    def try_acquire(self, count: int = 1) -> bool:
        """
        Try to acquire sandbox slots.

        Args:
            count: Number of slots to acquire

        Returns:
            True if slots were acquired, False if budget exhausted
        """
        with self._lock:
            if self.created + count > self.max:
                return False
            self.created += count
            self.active += count
            return True

    def release(self, count: int = 1) -> None:
        """
        Release sandbox slots (mark as no longer active).

        Note: This doesn't restore budget - created count stays the same.
        Budget tracks total sandboxes ever created, not concurrent ones.

        Args:
            count: Number of slots to release
        """
        with self._lock:
            self.active = max(0, self.active - count)

    @property
    def remaining(self) -> int:
        """Get remaining sandbox budget."""
        with self._lock:
            return self.max - self.created

    @property
    def status(self) -> BudgetStatus:
        """Get current budget status."""
        with self._lock:
            return BudgetStatus(
                max_sandboxes=self.max,
                created=self.created,
                active=self.active,
                remaining=self.max - self.created,
            )

    def can_acquire(self, count: int = 1) -> bool:
        """
        Check if count slots can be acquired without actually acquiring.

        Useful for checking batch feasibility before starting.

        Args:
            count: Number of slots to check

        Returns:
            True if slots are available
        """
        with self._lock:
            return self.created + count <= self.max

    def __repr__(self) -> str:
        status = self.status
        return (
            f"SandboxBudget(max={status.max_sandboxes}, "
            f"created={status.created}, active={status.active}, "
            f"remaining={status.remaining})"
        )
