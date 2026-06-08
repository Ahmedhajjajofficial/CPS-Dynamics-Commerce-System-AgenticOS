"""
Unit tests for CRDT implementations.
"""

import pytest
from crdt import (
    GCounter, PNCounter, ORSet, ORSetElement,
    LWWRegister, CRDTManager,
)


# ── GCounter ──────────────────────────────────────────────────────────────────


class TestGCounter:
    def test_initial_value_is_zero(self):
        c = GCounter("n1")
        assert c.value == 0

    def test_single_increment(self):
        c = GCounter("n1")
        c.increment(5)
        assert c.value == 5

    def test_default_increment_is_one(self):
        c = GCounter("n1")
        c.increment()
        assert c.value == 1

    def test_multiple_increments_accumulate(self):
        c = GCounter("n1")
        c.increment(3)
        c.increment(7)
        assert c.value == 10

    def test_negative_increment_raises(self):
        c = GCounter("n1")
        with pytest.raises(ValueError, match="positive"):
            c.increment(-1)

    def test_version_increments_on_update(self):
        c = GCounter("n1")
        assert c.version == 0
        c.increment(1)
        assert c.version == 1
        c.increment(1)
        assert c.version == 2

    def test_merge_takes_max_per_node(self):
        a = GCounter("n1")
        b = GCounter("n2")
        a.increment(5)
        b.increment(3)
        merged = a.merge(b)
        assert merged.value == 8  # 5 + 3

    def test_merge_same_node_takes_max(self):
        a = GCounter("n1")
        b = GCounter("n1")
        a.increment(5)
        b.increment(3)
        merged = a.merge(b)
        assert merged.value == 5  # max(5, 3)

    def test_merge_is_commutative(self):
        a = GCounter("n1")
        b = GCounter("n2")
        a.increment(5)
        b.increment(3)
        assert a.merge(b).value == b.merge(a).value

    def test_merge_is_idempotent(self):
        a = GCounter("n1")
        b = GCounter("n2")
        a.increment(5)
        b.increment(3)
        m1 = a.merge(b)
        m2 = m1.merge(b)
        assert m1.value == m2.value

    def test_serialization_roundtrip(self):
        c = GCounter("n1")
        c.increment(42)
        d = c.to_dict()
        restored = GCounter.from_dict(d)
        assert restored.value == 42
        assert restored.node_id == "n1"
        assert restored.version == c.version

    def test_to_dict_structure(self):
        c = GCounter("n1")
        c.increment(7)
        d = c.to_dict()
        assert d["type"] == "GCounter"
        assert d["node_id"] == "n1"
        assert d["increments"]["n1"] == 7


# ── PNCounter ─────────────────────────────────────────────────────────────────


class TestPNCounter:
    def test_initial_value_is_zero(self):
        c = PNCounter("n1")
        assert c.value == 0

    def test_increment(self):
        c = PNCounter("n1")
        c.increment(10)
        assert c.value == 10

    def test_decrement(self):
        c = PNCounter("n1")
        c.increment(10)
        c.decrement(3)
        assert c.value == 7

    def test_positive_and_negative_properties(self):
        c = PNCounter("n1")
        c.increment(10)
        c.decrement(3)
        assert c.positive == 10
        assert c.negative == 3

    def test_negative_increment_raises(self):
        c = PNCounter("n1")
        with pytest.raises(ValueError, match="decrement"):
            c.increment(-1)

    def test_negative_decrement_raises(self):
        c = PNCounter("n1")
        with pytest.raises(ValueError, match="increment"):
            c.decrement(-1)

    def test_merge_two_nodes(self):
        a = PNCounter("n1")
        b = PNCounter("n2")
        a.increment(10)
        a.decrement(2)
        b.increment(5)
        b.decrement(1)
        merged = a.merge(b)
        assert merged.value == 12  # (10+5) - (2+1)

    def test_merge_is_commutative(self):
        a = PNCounter("n1")
        b = PNCounter("n2")
        a.increment(10)
        b.decrement(3)
        assert a.merge(b).value == b.merge(a).value

    def test_merge_is_idempotent(self):
        a = PNCounter("n1")
        b = PNCounter("n2")
        a.increment(10)
        b.increment(5)
        m1 = a.merge(b)
        m2 = m1.merge(b)
        assert m1.value == m2.value

    def test_serialization_roundtrip(self):
        c = PNCounter("n1")
        c.increment(10)
        c.decrement(3)
        d = c.to_dict()
        restored = PNCounter.from_dict(d)
        assert restored.value == 7
        assert restored.node_id == "n1"

    def test_to_dict_structure(self):
        c = PNCounter("n1")
        c.increment(10)
        c.decrement(3)
        d = c.to_dict()
        assert d["type"] == "PNCounter"
        assert d["increments"]["n1"] == 10
        assert d["decrements"]["n1"] == 3

    def test_version_increments(self):
        c = PNCounter("n1")
        c.increment(1)
        assert c.version == 1
        c.decrement(1)
        assert c.version == 2


# ── ORSet ─────────────────────────────────────────────────────────────────────


class TestORSet:
    def test_empty_set(self):
        s = ORSet("n1")
        assert s.elements == set()

    def test_add_element(self):
        s = ORSet("n1")
        s.add("a")
        assert "a" in s.elements

    def test_add_returns_element_id(self):
        s = ORSet("n1")
        eid = s.add("a")
        assert isinstance(eid, str)
        assert len(eid) > 0

    def test_add_with_custom_id(self):
        s = ORSet("n1")
        eid = s.add("a", element_id="custom-id")
        assert eid == "custom-id"

    def test_remove_element(self):
        s = ORSet("n1")
        s.add("a")
        s.add("b")
        s.remove("a")
        assert "a" not in s.elements
        assert "b" in s.elements

    def test_remove_returns_removed_ids(self):
        s = ORSet("n1")
        s.add("a", element_id="id1")
        s.add("a", element_id="id2")
        removed = s.remove("a")
        assert set(removed) == {"id1", "id2"}

    def test_remove_nonexistent_returns_empty(self):
        s = ORSet("n1")
        removed = s.remove("nonexistent")
        assert removed == []

    def test_remove_by_id(self):
        s = ORSet("n1")
        eid = s.add("a")
        assert s.remove_by_id(eid) is True
        assert "a" not in s.elements

    def test_remove_by_id_nonexistent(self):
        s = ORSet("n1")
        assert s.remove_by_id("no-such-id") is False

    def test_remove_by_id_already_removed(self):
        s = ORSet("n1")
        eid = s.add("a")
        s.remove_by_id(eid)
        assert s.remove_by_id(eid) is False

    def test_contains(self):
        s = ORSet("n1")
        s.add("a")
        assert s.contains("a") is True
        assert s.contains("b") is False

    def test_removed_elements_property(self):
        s = ORSet("n1")
        s.add("a")
        s.add("b")
        s.remove("a")
        assert "a" in s.removed_elements
        assert "b" not in s.removed_elements

    def test_merge_disjoint_sets(self):
        a = ORSet("n1")
        b = ORSet("n2")
        a.add("x", element_id="e1")
        b.add("y", element_id="e2")
        merged = a.merge(b)
        assert merged.elements == {"x", "y"}

    def test_merge_preserves_removal(self):
        a = ORSet("n1")
        b = ORSet("n2")
        eid = a.add("x", element_id="shared")
        b_copy = ORSet("n2")
        b_copy.add("x", element_id="shared")
        b_copy.remove("x")
        merged = a.merge(b_copy)
        assert "x" not in merged.elements

    def test_merge_is_commutative(self):
        a = ORSet("n1")
        b = ORSet("n2")
        a.add("x", element_id="e1")
        b.add("y", element_id="e2")
        m1 = a.merge(b)
        m2 = b.merge(a)
        assert m1.elements == m2.elements

    def test_serialization_roundtrip(self):
        s = ORSet("n1")
        s.add("a", element_id="id1")
        s.add("b", element_id="id2")
        s.remove("a")
        d = s.to_dict()
        restored = ORSet.from_dict(d)
        assert restored.elements == {"b"}
        assert restored.node_id == "n1"

    def test_version_increments(self):
        s = ORSet("n1")
        assert s.version == 0
        s.add("a")
        assert s.version == 1
        s.remove("a")
        assert s.version == 2


class TestORSetElement:
    def test_serialization_roundtrip(self):
        elem = ORSetElement(
            element_id="e1",
            value="product_A",
            is_removed=False,
            added_at="2024-01-01T00:00:00",
        )
        d = elem.to_dict()
        restored = ORSetElement.from_dict(d)
        assert restored.element_id == "e1"
        assert restored.value == "product_A"
        assert restored.is_removed is False

    def test_removed_element(self):
        elem = ORSetElement(
            element_id="e1",
            value="product_A",
            is_removed=True,
            added_at="2024-01-01T00:00:00",
            removed_at="2024-01-02T00:00:00",
        )
        d = elem.to_dict()
        restored = ORSetElement.from_dict(d)
        assert restored.is_removed is True
        assert restored.removed_at == "2024-01-02T00:00:00"


# ── LWWRegister ───────────────────────────────────────────────────────────────


class TestLWWRegister:
    def test_initial_value_empty(self):
        r = LWWRegister("n1")
        assert r.value == ""

    def test_set_and_get(self):
        r = LWWRegister("n1")
        r.set("hello")
        assert r.value == "hello"

    def test_set_with_explicit_timestamp(self):
        r = LWWRegister("n1")
        r.set("old", timestamp="2024-01-01T00:00:00")
        r.set("new", timestamp="2024-01-02T00:00:00")
        assert r.value == "new"

    def test_older_timestamp_does_not_overwrite(self):
        r = LWWRegister("n1")
        r.set("new", timestamp="2024-01-02T00:00:00")
        r.set("old", timestamp="2024-01-01T00:00:00")
        assert r.value == "new"

    def test_merge_takes_latest(self):
        a = LWWRegister("n1")
        b = LWWRegister("n2")
        a.set("old", timestamp="2024-01-01T00:00:00")
        b.set("new", timestamp="2024-01-02T00:00:00")
        merged = a.merge(b)
        assert merged.value == "new"

    def test_merge_is_commutative(self):
        a = LWWRegister("n1")
        b = LWWRegister("n2")
        a.set("val_a", timestamp="2024-01-01T00:00:00")
        b.set("val_b", timestamp="2024-01-02T00:00:00")
        assert a.merge(b).value == b.merge(a).value

    def test_serialization_roundtrip(self):
        r = LWWRegister("n1")
        r.set("hello", timestamp="2024-06-01T12:00:00")
        d = r.to_dict()
        restored = LWWRegister.from_dict(d)
        assert restored.value == "hello"
        assert restored.timestamp == "2024-06-01T12:00:00"
        assert restored.node_id == "n1"

    def test_version_increments(self):
        r = LWWRegister("n1")
        r.set("a")
        assert r.version == 1
        r.set("b")
        assert r.version == 2

    def test_timestamp_property(self):
        r = LWWRegister("n1")
        r.set("val", timestamp="2024-06-01T12:00:00")
        assert r.timestamp == "2024-06-01T12:00:00"


# ── CRDTManager ───────────────────────────────────────────────────────────────


class TestCRDTManager:
    def test_create_pn_counter(self):
        mgr = CRDTManager("n1")
        c = mgr.create_counter("inv_001", "PN")
        assert isinstance(c, PNCounter)
        assert mgr.get("inv_001") is c

    def test_create_g_counter(self):
        mgr = CRDTManager("n1")
        c = mgr.create_counter("sales_001", "G")
        assert isinstance(c, GCounter)

    def test_create_orset(self):
        mgr = CRDTManager("n1")
        s = mgr.create_orset("promotions")
        assert isinstance(s, ORSet)

    def test_create_lww_register(self):
        mgr = CRDTManager("n1")
        r = mgr.create_lww_register("price_001")
        assert isinstance(r, LWWRegister)

    def test_get_nonexistent_returns_none(self):
        mgr = CRDTManager("n1")
        assert mgr.get("nope") is None

    def test_merge_existing(self):
        mgr = CRDTManager("n1")
        local = mgr.create_counter("inv", "PN")
        local.increment(5)

        remote = PNCounter("n2")
        remote.increment(3)

        merged = mgr.merge("inv", remote)
        assert merged.value == 8

    def test_merge_new_crdt(self):
        mgr = CRDTManager("n1")
        remote = GCounter("n2")
        remote.increment(10)
        merged = mgr.merge("new_counter", remote)
        assert merged.value == 10
        assert mgr.get("new_counter") is not None

    def test_get_all_states(self):
        mgr = CRDTManager("n1")
        mgr.create_counter("c1", "G")
        mgr.create_orset("s1")
        states = mgr.get_all_states()
        assert "c1" in states
        assert "s1" in states
        assert states["c1"]["type"] == "GCounter"
        assert states["s1"]["type"] == "ORSet"

    def test_load_states(self):
        mgr1 = CRDTManager("n1")
        c = mgr1.create_counter("c1", "G")
        c.increment(42)
        states = mgr1.get_all_states()

        mgr2 = CRDTManager("n2")
        mgr2.load_states(states)
        restored = mgr2.get("c1")
        assert restored is not None
        assert restored.value == 42

    def test_load_states_all_types(self):
        mgr1 = CRDTManager("n1")
        gc = mgr1.create_counter("gc", "G")
        gc.increment(5)
        pn = mgr1.create_counter("pn", "PN")
        pn.increment(10)
        pn.decrement(3)
        orset = mgr1.create_orset("os")
        orset.add("x")
        lww = mgr1.create_lww_register("lww")
        lww.set("val")

        states = mgr1.get_all_states()
        mgr2 = CRDTManager("n2")
        mgr2.load_states(states)

        assert mgr2.get("gc").value == 5
        assert mgr2.get("pn").value == 7
        assert "x" in mgr2.get("os").elements
        assert mgr2.get("lww").value == "val"
