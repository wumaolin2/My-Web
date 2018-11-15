package controllers

import (
	"github.com/astaxie/beego"
	"path"
	"time"
	"github.com/astaxie/beego/orm"
	"newsWeb/models"
	"math"
	"strconv"
)

type ArticleController struct {
	beego.Controller
}

//展示文章列表页
func (this *ArticleController) ShowArticleList() {
	userName := this.GetSession("userName")
	if userName == nil{
		this.Redirect("/ShowLogin",302)
		return
	}
     this.Data["userName"]=userName
	//查询数据库，拿出数据，传递给视图
	//获取orm对象
	o := orm.NewOrm()
	//获取查询对象
	var articles []models.Article
	//查询
	//queryseter  高级查询使用的数据类型
	qs := o.QueryTable("Article")
	//查询所有的文章
	//qs.All(&articles)//select * from article

	//实现分页
	//获取总记录数和总页数
	count, _ := qs.Count()

	pageSize := int64(2)

	pageCount := float64(count) / float64(pageSize)

	pageCount = math.Ceil(pageCount)

	//向上取整
	//把数据传递给视图
	this.Data["count"] = count
	this.Layout="layout.html"
	this.Data["pageCount"] = pageCount

	//获取首页末页数据
	pageIndex, err := this.GetInt("pageIndex")
	if err != nil {
		pageIndex = 1
	}
	//获取分页的数据
	start := pageSize * (int64(pageIndex) - 1 )
	qs.Limit(pageSize, start).RelatedSel().All(&articles)

	//根据传递的类型获取相应的文章
	//获取数据
	typeName:=this.GetString("select")
	this.Data["typeName"]=typeName
	qs.Limit(pageSize, start).RelatedSel().Filter("ArticleType__TypeName",typeName).All(&articles)



	var articleTypes []models.ArticleType
	o.QueryTable("ArticleType").All(&articleTypes)
	this.Data["articleTypes"]=articleTypes
	this.Data["pageIndex"] = pageIndex
	this.Data["articles"] = articles
	this.TplName = "index.html"
}

//展示添加文章页面
func (this *ArticleController) ShowAddArticle() {
	userName := this.GetSession("userName")
	if userName == nil{
		this.Redirect("/ShowLogin",302)
		return
	}
	this.Data["userName"]=userName
	//获取orm对象
	o:=orm.NewOrm()
	//定以个容器存储类型数据
	var articleTypes []models.ArticleType
	//huoqu 获取类型并方到容器里面
	o.QueryTable("ArticleType").All(&articleTypes)
	//展示给前端
	this.Data["articleTypes"]=articleTypes
	this.Layout="layout.html"
	this.TplName = "add.html"
}

//处理添加文章业务
func (this *ArticleController) HandeAddArticle() {
	//接受数据
	artileName := this.GetString("articleName")
	content := this.GetString("content")
	//校验数据
	if artileName == "" || content == "" {
		this.Data["errmsg"] = "文章标题或内容不能为空"
		this.TplName = "add.html"
		return
	}

	//接收图片
	file, head, err := this.GetFile("uploadname")
	if err != nil {
		this.Data["errmsg"] = "获取文件失败"
		this.TplName = "add.html"
		return
	}
	defer file.Close()
	//1.判断文件大小
	if head.Size > 500000 {
		this.Data["errmsg"] = "文件太大，上传失败"
		this.TplName = "add.html"
		return
	}

	//2.判断图片格式
	//1.jpg
	fileExt := path.Ext(head.Filename)
	if fileExt != ".jpg" && fileExt != ".png" && fileExt != ".jpeg" {
		this.Data["errmsg"] = "文件格式不正确，请重新上传"
		this.TplName = "add.html"
		return
	}

	//3.文件名防止重复
	fileName := time.Now().Format("2006-01-02-15-04-05") + fileExt
	this.SaveToFile("uploadname", "./static/image/"+fileName)

	//处理数据
	//数据库的插入操作
	//获取orm对象
	o := orm.NewOrm()
	//获取插入对象
	var article models.Article
	var articleType models.ArticleType
	//给插入对象赋值
	//获取文章类型
	TypeName:=this.GetString("select")
	//给文章类型赋值
	articleType.TypeName=TypeName
	//查询
	o.Read(&articleType,"TypeName")
	article.ArticleType=&articleType
	article.Title = artileName
	article.Content = content
	article.Image = "/static/image/" + fileName

	//插入
	_, err = o.Insert(&article)
	if err != nil {
		this.Data["errmsg"] = "添加文章失败，请重新添加"
		this.TplName = "add.html"
		return
	}
	//返回页面
	this.Redirect("/article/articleList", 302)

}

//展示文章详细内容
func (this *ArticleController) ShowArticleDetail() {
	//获取数据
	articleId, err := this.GetInt("Id")
	//校验数据
	if err != nil {
		this.Data["errmsg"] = " 请求路径错误1"
		this.TplName = "index.html"
		return
	}
	//处理数据
	//查询
	//获取orm对象
	o := orm.NewOrm()
	//获取查询对象
	var article models.Article
	//给查询对条件赋值
	article.Id = articleId
	//查询
	err = o.Read(&article, "Id")
	if err != nil {
		this.Data["errmsg"] = "请求路径错误2"
		this.TplName = "index.html"
		return
	}
	//获取article对象
	//上面已经获取
	//获取多对多操作对象
	m2m:=o.QueryM2M(&article,"Users")
    // 获取要插入的数据
    //获取user表
    var user models.User
    //huoqu获取存储在session中的用户名
    userName :=this.GetSession("userName")
    this.Data["userName"]=userName
    //把用户名给表中的userName
	user.UserName=userName.(string)
	o.Read(&user,"UserName")
	//插入多对多关系
	m2m.Add(user)
	//第一种多对多的查询
	o.LoadRelated(&article,"ArticleType")
	//第二种多对多的查询
	var users []models.User
	o.QueryTable("User").Filter("Article__Article__Id",articleId).Distinct().All(&users)
	//返回数据
	this.Data["users"]=users
	// 返回数据
	this.Data["article"] = article
	this.Layout="layout.html"
	this.TplName = "content.html"
}

//显示编辑文章详情页

func (this *ArticleController) ShowUpdateArticle() {
	userName := this.GetSession("userName")
	if userName == nil{
		this.Redirect("/ShowLogin",302)
		return
	}
	this.Data["userName"]=userName

	//获取数据
	ArticleId, err := this.GetInt("Id")

	//校验数据
	if err != nil {
		beego.Error("请求路径错误")
		this.Redirect("/article/articleList?errmsg", 302)
	}
	//处理数据
	//查询
	o := orm.NewOrm()
	var article models.Article
	article.Id = ArticleId
	o.Read(&article)
	//返回数据
	this.Data["article"] = article
	this.Layout="layout.html"
	this.TplName = "update.html"
}

//	文件上传函数
func UploadFile(this *ArticleController, filePath string) string {
	//	接受图片
	file, head, err := this.GetFile("uploadname")
	if err != nil {
		this.Data["errmsg"] = "获取文件失败"
		this.TplName = "add.html"
		return ""
	}
	defer file.Close()
	//1.判断文件大小
	if head.Size > 500000 {
		this.Data["errmsg"] = "文件太大，上传失败"
		this.TplName = "add.html"
		return ""
	}

	//2.判断图片格式
	//1.jpg
	fileExt := path.Ext(head.Filename)
	if fileExt != ".jpg" && fileExt != ".png" && fileExt != ".jpeg" {
		this.Data["errmsg"] = "文件格式不正确，请重新上传"
		this.TplName = "add.html"
		return ""
	}

	//3.文件名防止重复
	fileName := time.Now().Format("2006-01-02-15-04-05") + fileExt
	this.SaveToFile("uploadname", "./static/image/"+fileName)
	return "./static/image/" + fileName
}

//处理编辑文章业务
func (this *ArticleController) HandleUpdateArticle() {
	//	获取数据
	articleName := this.GetString("articleName")
	content := this.GetString("content")
	fileName := UploadFile(this, "uploadname")
	articleId, err2 := this.GetInt("id")
	//	硝烟数据
	if articleName == "" || content == "" || fileName == "" || err2 != nil {
		errmsg := "内容不能为空"
		this.Redirect("/article/updateArticle?id="+strconv.Itoa(articleId)+"&errmsg="+errmsg, 302)
		return
	}

	//处理数据
	//更新操作 update
	//获取orm对象
	o := orm.NewOrm()
	//获取更新对象
	var article models.Article
	//给更新对象赋值
	article.Id = articleId
	err := o.Read(&article)
	if err != nil {
		errmsg := "更新的文章不存在"
		this.Redirect("/article/updateArticle?id="+strconv.Itoa(articleId)+"&errmsg="+errmsg, 302)
		return
	}
	//给更新对象赋值
	article.Title = articleName
	article.Content = content
	article.Image = fileName
	//更新
	o.Update(&article)
	//返回数据
	this.Redirect("/article/articleList", 302)
}
//删除操作
func(this*ArticleController)DeleteArticle(){
//	获取shuju
   articleId,err:= this.GetInt("Id")
	if err!=nil  {

		beego.Error("路径错误")
		this.Redirect("/article/articleList",302)
		return
		}
	//处理数据
	//删除操作
	//获取orm对象
	o := orm.NewOrm()
	//获取删除对象
	var article models.Article
	//给删除对象赋值000000000000000
	article.Id = articleId
	//删除
	_,err = o.Delete(&article)
	if err != nil{
		beego.Error("删除失败")
		this.Redirect("/article/articleList",302)
		return
	}

	//返回数据
	this.Redirect("/article/articleList",302)
}


//展示添加类型页面
func(this*ArticleController)ShowAddType(){
	userName := this.GetSession("userName")
	if userName == nil{
		this.Redirect("/ShowLogin",302)
		return
	}
	this.Data["userName"]=userName
	//获取数据所有类型数据并展
	//huoqu orm对象
	o:=orm.NewOrm()
	//查询容器
	var articleTypes []models.ArticleType
	//指定查询表
	qs:=o.QueryTable("ArticleType")
	qs.All(&articleTypes)
	//返回数据给试图
	this.Data["articleTypes"]=articleTypes
	this.Layout="layout.html"
	this.TplName="addType.html"
}


//处理类型添加业务
func(this*ArticleController)HandleAddType(){
//	获取数据
 typeName:=this.GetString("typeName")
//硝烟数据
	if typeName=="" {
		errmsg:="类型名不能为空"
        this.Redirect("/article/addType?&errmsg="+errmsg,301)
return
	}
//chuli 数据
//插入错作
//获取orm对乡
o:=orm.NewOrm()
//获取插入对象
var articleType models.ArticleType

//给charu 插入对象赋值 T
articleType.TypeName=typeName

//插入
   _,err:= o.Insert(&articleType)
if err!=nil {
	beego.Error(err)
	errmsg:="插入失败，清重新插入"
	this.Redirect("/article/addType?&errmsg="+errmsg,302)
	return
}

//返回数据
this.Redirect("/article/addType",30)
}

//删除类型业务
func (this*ArticleController)ShowDeleteType()  {
//	获取数据
  typeId,err:= this.GetInt("Id")
//硝烟数据
  if err!=nil {
	errmsg:="删除失败1"
	this.Redirect("/article/addType?errmsg="+errmsg,302)
	return
}

// 处理数据

 //获取orm对象
 o:= orm.NewOrm()
 //获取删除对象
 var articleType models.ArticleType
 //gei 给删除对象赋值
 articleType.Id = typeId
 //删除
  _  ,err =o.Delete(&articleType)
   if err !=nil {
   	errmsg := "删除失败2 "
   	this.Redirect("/article/addType?errmsg="+errmsg,302)
   	return
   }
   //返回数据
   this.Redirect("/article/addType",302)


  }
