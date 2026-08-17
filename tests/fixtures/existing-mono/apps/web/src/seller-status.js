export function sellerStatusLabel(status) {
  const labels = { active: "正常", review: "待审核", suspended: "已暂停" };
  return labels[status] ?? "未知";
}
