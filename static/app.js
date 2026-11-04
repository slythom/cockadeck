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

// Confirmation dialog: replaces hx-confirm's native popup with #confirm-dialog.
// htmx fires htmx:confirm when a request carries hx-confirm; cancelling it lets us
// resume the request via detail.issueRequest() or abort via detail.dropRequest().
const confirmDialog = document.getElementById("confirm-dialog");
const confirmMsg = confirmDialog.querySelector(".confirm-msg");
const confirmOk = confirmDialog.querySelector("[data-confirm-ok]");
const confirmCancel = confirmDialog.querySelector("[data-confirm-cancel]");
let settleConfirm = null;

function closeConfirm(result) {
	if (!settleConfirm) return;
	const resolve = settleConfirm;
	settleConfirm = null;
	confirmDialog.classList.add("closing");
	setTimeout(() => {
		confirmDialog.close();
		confirmDialog.classList.remove("closing");
	}, 120);
	resolve(result);
}

function askConfirm(message, okLabel, subject) {
	const i = subject ? message.indexOf("%s") : -1;
	if (i < 0) {
		confirmMsg.textContent = message;
	} else {
		const strong = document.createElement("strong");
		strong.textContent = subject;
		confirmMsg.replaceChildren(message.slice(0, i), strong, message.slice(i + 2));
	}
	confirmOk.textContent = okLabel;
	confirmDialog.showModal();
	confirmCancel.focus();
	return new Promise((resolve) => {
		settleConfirm = resolve;
	});
}

confirmOk.addEventListener("click", () => closeConfirm(true));
confirmCancel.addEventListener("click", () => closeConfirm(false));
confirmDialog.addEventListener("cancel", (e) => {
	e.preventDefault();
	closeConfirm(false);
});
confirmDialog.addEventListener("close", () => closeConfirm(false));
confirmDialog.addEventListener("click", (e) => {
	const r = confirmDialog.getBoundingClientRect();
	if (e.clientX < r.left || e.clientX > r.right || e.clientY < r.top || e.clientY > r.bottom) {
		closeConfirm(false);
	}
});

document.addEventListener("htmx:confirm", (e) => {
	const src = e.target.closest("[hx-confirm]");
	if (!src) return;
	e.preventDefault();
	askConfirm(src.getAttribute("hx-confirm"), src.dataset.confirmLabel || "Confirm", src.dataset.confirmSubject).then((ok) => {
		if (ok) {
			e.detail.issueRequest();
		} else {
			e.detail.dropRequest();
		}
	});
});
