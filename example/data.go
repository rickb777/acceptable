package main

type Declaration struct {
	Proclamation string
	Articles     []Article
}

type Article struct {
	N    int
	Text string
}

var en = Declaration{
	Proclamation: "The General Assembly proclaims this Universal Declaration of Human Rights as a common standard of achievement for all peoples and all nations, to the end that every individual and every organ of society, keeping this Declaration constantly in mind, shall strive by teaching and education to promote respect for these rights and freedoms and by progressive measures, national and international, to secure their universal and effective recognition and observance, both among the peoples of Member States themselves and among the peoples of territories under their jurisdiction.",

	Articles: []Article{
		{1, "All human beings are born free and equal in dignity and rights. They are endowed with reason and conscience and should act towards one another in a spirit of brotherhood."},

		{2, "Everyone is entitled to all the rights and freedoms set forth in this Declaration, without distinction of any kind, such as race, colour, sex, language, religion, political or other opinion, national or social origin, property, birth or other status. Furthermore, no distinction shall be made on the basis of the political, jurisdictional or international status of the country or territory to which a person belongs, whether it be independent, trust, non-self-governing or under any other limitation of sovereignty."},

		{3, "Everyone has the right to life, liberty and security of person."},

		{4, "No one shall be held in slavery or servitude; slavery and the slave trade shall be prohibited in all their forms."},

		{5, "No one shall be subjected to torture or to cruel, inhuman or degrading treatment or punishment."},
	},
}

var fr = Declaration{
	Proclamation: "L'Assemblée générale proclame la présente Déclaration universelle des droits de l'homme comme l'idéal commun à atteindre par tous les peuples et toutes les nations afin que tous les individus et tous les organes de la société, ayant cette Déclaration constamment à l'esprit, s'efforcent, par l'enseignement et l'éducation, de développer le respect de ces droits et libertés et d'en assurer, par des mesures progressives d'ordre national et international, la reconnaissance et l'application universelles et effectives, tant parmi les populations des Etats Membres eux-mêmes que parmi celles des territoires placés sous leur juridiction.",

	Articles: []Article{
		{1, "Tous les êtres humains naissent libres et égaux en dignité et en droits. Ils sont doués de raison et de conscience et doivent agir les uns envers les autres dans un esprit de fraternité."},

		{2, "1. Chacun peut se prévaloir de tous les droits et de toutes les libertés proclamés dans la présente Déclaration, sans distinction aucune, notamment de race, de couleur, de sexe, de langue, de religion, d'opinion politique ou de toute autre opinion, d'origine nationale ou sociale, de fortune, de naissance ou de toute autre situation.\n2. De plus, il ne sera fait aucune distinction fondée sur le statut politique, juridique ou international du pays ou du territoire dont une personne est ressortissante, que ce pays ou territoire soit indépendant, sous tutelle, non autonome ou soumis à une limitation quelconque de souveraineté."},

		{3, "Tout individu a droit à la vie, à la liberté et à la sûreté de sa personne."},

		{4, "Nul ne sera tenu en esclavage ni en servitude; l'esclavage et la traite des esclaves sont interdits sous toutes leurs formes."},

		{5, "Nul ne sera soumis à la torture, ni à des peines ou traitements cruels, inhumains ou dégradants."},
	},
}

var es = Declaration{
	Proclamation: "como ideal común por el que todos los pueblos y naciones deben esforzarse, a fin de que tanto los individuos como las instituciones, inspirándose constantemente en ella, promuevan, mediante la enseñanza y la educación, el respeto a estos derechos y libertades, y aseguren, por medidas progresivas de carácter nacional e internacional, su reconocimiento y aplicación universales y efectivos, tanto entre los pueblos de los Estados Miembros como entre los de los territorios colocados bajo su jurisdicción.",

	Articles: []Article{
		{1, "Todos los seres humanos nacen libres e iguales en dignidad y derechos y, dotados como están de razón y conciencia, deben comportarse fraternalmente los unos con los otros."},

		{2, "Toda persona tiene todos los derechos y libertades proclamados en esta Declaración, sin distinción alguna de raza, color, sexo, idioma, religión, opinión política o de cualquier otra índole, origen nacional o social, posición económica, nacimiento o cualquier otra condición. Además, no se hará distinción alguna fundada en la condición política, jurídica o internacional del país o territorio de cuya jurisdicción dependa una persona, tanto si se trata de un país independiente, como de un territorio bajo administración fiduciaria, no autónomo o sometido a cualquier otra limitación de soberanía."},

		{3, "Todo individuo tiene derecho a la vida, a la libertad y a la seguridad de su persona."},

		{4, "Nadie estará sometido a esclavitud ni a servidumbre, la esclavitud y la trata de esclavos están prohibidas en todas sus formas."},

		{5, "Nadie será sometido a torturas ni a penas o tratos crueles, inhumanos o degradantes."},
	},
}

var ru = Declaration{
	Proclamation: "провозглашает настоящую Всеобщую декларацию прав человека в качестве задачи, к выполнению которой должны стремиться все народы и государства с тем, чтобы каждый человек и каждый орган общества, постоянно имея в виду настоящую Декларацию, стремились путем просвещения и образования содействовать уважению этих прав и свобод и обеспечению, путем национальных и международных прогрессивных мероприятий, всеобщего и эффективного признания и осуществления их как среди народов государств-членов Организации, так и среди народов территорий, находящихся под их юрисдикцией.",

	Articles: []Article{
		{1, "Все люди рождаются свободными и равными в своем достоинстве и правах. Они наделены разумом и совестью и должны поступать в отношении друг друга в духе братства."},

		{2, "Каждый человек должен обладать всеми правами и всеми свободами, провозглашенными настоящей Декларацией, без какого бы то ни было различия, как-то в отношении расы, цвета кожи, пола, языка, религии, политических или иных убеждений, национального или социального происхождения, имущественного, сословного или иного положения. 	Кроме того, не должно проводиться никакого различия на основе политического, правового или международного статуса страны или территории, к которой человек принадлежит, независимо от того, является ли эта территория независимой, подопечной, несамоуправляющейся или как-либо иначе ограниченной в своем суверенитете."},

		{3, "Каждый человек имеет право на жизнь, на свободу и на личную неприкосновенность."},

		{4, "Никто не должен содержаться в рабстве или в подневольном состоянии; рабство и работорговля запрещаются во всех их видах."},

		{5, "Никто не должен подвергаться пыткам или жестоким, бесчеловечным или унижающим его достоинство обращению и наказанию."},
	},
}
