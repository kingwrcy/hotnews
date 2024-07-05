window.__unocss = {
    theme: {
        colors: {
            fg: "rgba(51, 51, 51,.7)",
            hover: "#5468ff",
            link: "#009ff7",
        },
    },
    shortcuts: {
        btn: "bg-#5e7ce0 px-4 py-0.5 text-white outline-0 rounded text-sm hover:bg-#7693f5 border-0 cursor-pointer",
        input: "focus:border-#5e7ce0 border rounded px-2 py-1 outline-0 text-sm border-#e5e7eb border-solid",
        aLink: "underline text-link mx-2",
        bLink: "text-link mx-1 hover:text-gray",
        'x-post-tag':'text-xs shadow border px-1.5 py-0.5 rounded b-solid b-coolGray cursor-pointer hover:opacity-80',
        tag1:'bg-white text-gray-500',
    },
}

$(function () {
    $("input[name='result']").click(function () {
        const val = $(this).val()
        const index = $(this).data('index')
        if (val === 'reject') {
            $(`#reason-${index}`).show()
        } else {
            $(`#reason-${index}`).hide()
        }
    })

    const $sidebar = $("#sidebar");
    $("#showSidebar").click((e)=>{
        $sidebar.show().fadeIn()
        e.stopPropagation()
    })

    $(document.body).click(()=>{
      if ($sidebar.css('display') === 'block'){
          $sidebar.hide().fadeOut()
      }
    })
})