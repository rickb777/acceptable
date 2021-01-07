package main

type Declaration struct {
	Title    string
	Articles []Article
}

type Article struct {
	N    int
	Text string
}

var en = Declaration{
	Title: "a common standard of achievement for all peoples and all nations, to the end that every individual and every organ of society, keeping this Declaration constantly in mind, shall strive by teaching and education to promote respect for these rights and freedoms and by progressive measures, national and international, to secure their universal and effective recognition and observance",
	Articles: []Article{
		{1, "All human beings are born free and equal in dignity and rights. They are endowed with reason and conscience and should act towards one another in a spirit of brotherhood."},

		{2, "Everyone is entitled to all the rights and freedoms set forth in this Declaration, without distinction of any kind, such as race, colour, sex, language, religion, political or other opinion, national or social origin, property, birth or other status. Furthermore, no distinction shall be made on the basis of the political, jurisdictional or international status of the country or territory to which a person belongs, whether it be independent, trust, non-self-governing or under any other limitation of sovereignty."},

		{3, "Everyone has the right to life, liberty and security of person."},

		{4, "No one shall be held in slavery or servitude; slavery and the slave trade shall be prohibited in all their forms."},

		{5, "No one shall be subjected to torture or to cruel, inhuman or degrading treatment or punishment."},
	},
}

var fr = Declaration{
	Title: "constamment à l'esprit, s'efforcent, par l'enseignement et l'éducation, de développer le respect de ces droits et libertés et d'en assurer, par des mesures progressives d'ordre national et international, la reconnaissance et l'application universelles et effectives",
	Articles: []Article{
		{1, "Tous les êtres humains naissent libres et égaux en dignité et en droits. Ils sont doués de raison et de conscience et doivent agir les uns envers les autres dans un esprit de fraternité."},

		{2, "1. Chacun peut se prévaloir de tous les droits et de toutes les libertés proclamés dans la présente Déclaration, sans distinction aucune, notamment de race, de couleur, de sexe, de langue, de religion, d'opinion politique ou de toute autre opinion, d'origine nationale ou sociale, de fortune, de naissance ou de toute autre situation.\n2. De plus, il ne sera fait aucune distinction fondée sur le statut politique, juridique ou international du pays ou du territoire dont une personne est ressortissante, que ce pays ou territoire soit indépendant, sous tutelle, non autonome ou soumis à une limitation quelconque de souveraineté."},

		{3, "Tout individu a droit à la vie, à la liberté et à la sûreté de sa personne."},

		{4, "Nul ne sera tenu en esclavage ni en servitude; l'esclavage et la traite des esclaves sont interdits sous toutes leurs formes."},

		{5, "Nul ne sera soumis à la torture, ni à des peines ou traitements cruels, inhumains ou dégradants."},
	},
}
