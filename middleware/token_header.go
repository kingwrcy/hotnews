package middleware

//func CheckHeader(injector *do.Injector) gin.HandlerFunc {
//
//srv := do.MustInvoke[*service.TokenService](injector)
//return func(c *gin.Context) {
//	token := c.GetHeader("token")
//	if token == "" {
//		c.JSON(200, gin.H{
//			"code": 1,
//			"msg":  "token为空",
//		})
//		c.Abort()
//		return
//	}
//	if !srv.CheckExists(token) {
//		c.JSON(200, gin.H{
//			"code": 1,
//			"msg":  "token无效",
//		})
//		c.Abort()
//		return
//	}
//	c.Next()
//}
//
//}
