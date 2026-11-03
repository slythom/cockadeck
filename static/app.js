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
	input.dispatchEvent(new Event("change", { bubbles: true }));
});

function closeDropdowns() {
	document.querySelectorAll(".dropdown.open").forEach((dd) => {
		dd.classList.remove("open");
		dd.querySelector(".dropdown-btn").setAttribute("aria-expanded", "false");
	});
}

// Dropdowns: .dropdown-btn toggles .dropdown-menu; .dropdown-item sets the hidden input.
document.addEventListener("click", (e) => {
	const item = e.target.closest(".dropdown-item");
	if (item) {
		const dd = item.closest(".dropdown");
		const input = dd.querySelector("input[type=hidden]");
		input.value = item.dataset.value;
		dd.querySelector(".dropdown-label").textContent = item.textContent;
		dd.querySelectorAll(".dropdown-item.selected").forEach((el) => el.classList.remove("selected"));
		item.classList.add("selected");
		closeDropdowns();
		input.dispatchEvent(new Event("change", { bubbles: true }));
		return;
	}
	const btn = e.target.closest(".dropdown-btn");
	if (btn) {
		const dd = btn.closest(".dropdown");
		const open = dd.classList.toggle("open");
		btn.setAttribute("aria-expanded", String(open));
		return;
	}
	closeDropdowns();
});

document.addEventListener("keydown", (e) => {
	if (e.key === "Escape") closeDropdowns();
});
