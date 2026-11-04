// Pill steppers: .num-btn inside .num-input steps the sibling number input.
document.addEventListener("click", (e) => {
	const btn = e.target.closest(".num-btn");
	if (!btn) return;
	const input = btn.closest(".num-input")?.querySelector("input[type=number]");
	if (!input) return;
	if (Number(btn.dataset.step) < 0) {
		input.stepDown();
	} else {
		input.stepUp();
	}
	input.dispatchEvent(new Event("input", { bubbles: true }));
});
