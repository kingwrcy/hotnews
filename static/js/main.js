$(function (){
    $("input[name='result']").click(function (){
        const val = $(this).val()
        const index = $(this).data('index')
        if (val === 'reject'){
            $(`#reason-${index}`).show()
        }else{
            $(`#reason-${index}`).hide()
        }
    })

    // $("#approve").click(function (){
    //     const parent = $(this).parents("#approve-form");
    //     const reason = parent.find("textarea[name='reason']");
    //     const result = parent.find("input[name='inspect-result']");
    //     console.log(reason,result)
    //
    //     $.ajax({
    //         type: "POST",
    //         url: "/inspect",
    //         data: JSON.stringify({
    //             reason:reason.val(),result:result.val(),
    //             inspectType:'POST',
    //             postID:parent.data('post-id')
    //         }),
    //         success: function(){
    //
    //         },
    //         contentType: "application/json",
    //     });
    // })
})