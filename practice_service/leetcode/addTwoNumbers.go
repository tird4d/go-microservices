package leetcode

type ListNode struct {
	Val  int
	Next *ListNode
}

func LinkListGenerator(list []int) *ListNode {

	a := ListNode{}
	node := &a
	for i := 0; i < len(list); i++ {
		node = addNode(list[i], node)
	}

	return a.Next
}

func addNode(val int, node *ListNode) *ListNode {

	node.Val = val
	node.Next = &ListNode{}

	return node.Next
}

func AddTwoNumbers(l1 *ListNode, l2 *ListNode) *ListNode {

	return l1
}
