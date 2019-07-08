# A basic example in order to get the big picture

Roughly said, `manala` allows you to embed distributed templates in your projects and ease the synchronization of your projects when the reference templates are updated.

In this usage example, we are gonna implement a very basic feature, yet rich enough to measure the benefits of `manala` and to fully understand the basic concepts behind it.

## The scenario of our example

All your company's projects use [`PHP-CS-fixer`](https://github.com/FriendsOfPHP/PHP-CS-Fixer) in order to define your coding rules and apply them.

Your company would like to apply always the same coding rules on all of its projects, but maintaining the same set of rules in every project can be tedious and error-prone. In an idealistic world, the coding rules should be maintained in one place and passed on to all your projects as seamlessly as possible.

That's where manala enters the game ...

## Some wording: `project` vs `template` vs `repository`

In manala's vocabulary, your projects (the company's PHP projects in our example) are called ... `projects`.

In our example, our reference coding rules will be stored in a single place where they will be maintained. The file containing your coding rules is called a `template`. All the templates you maintain are made accessible to your colleagues through a `repository`.

## First step : install manala

```
curl -sL https://github.com/nervo/manala/raw/master/install.sh | sudo sh
```

> :bulb: Run `manala` in a console/terminal.

## Create your template repository and your first template

> :bulb: manala ships with some templates by default. Run `manala list` to display the list of available templates.

But in this example, we are gonna create our own template repository to better understand how manala works under the hood and enable you to develop your own templates when the need arises.

Run the following command to create your template repository : 

`mkdir ~/my-manala-template-repository`

Within this repository, we are gonna create a template group that will host our php rule template: 

`mkdir mkdir ~/my-manala-template-repository/my-php-templates`

> :bulb: template `repository` and `group` ? In manala's philosophy, a repository is viewed as a company-wide repository where you can store templates for various purposes and many profiles: infrastructure templates, backend developers, frontend developers, etc. In fact, your projects will not embed all the company's templates but just the subset of templates that's useful for your project. In our example, we are gonna embed only the templates under `my-php-templates` in our PHP projects.

Let's create a `.manala.yaml` file under the `my-php-templates` : 

```shell
  cd ~/my-manala-template-repository/my-php-templates
  touch ./.manala.yaml
```

> :bulb: the `.manala.yaml` acts as a manifest for your template group. It holds the name of your template group and indicates where its templates lie.

Now edit this file and put the following content : 

```yaml
manala:
    description: My company's PHP templates
    sync:
        - .manala
```

Now we are gonna create the `.manala` folder where all our PHP templates will be hosted : 

`mkdir ./.manala`

And finally, our PHP rule template:

`touch `./.manala/php-cs-rules.php`

And paste the following content:

```php
<?php

$header = <<<'EOF'
This file is part of the XXX project.

Copyright © My Company

@author My Company <contact@my-company.com>
EOF;

return [
    '@Symfony' => true,
    'psr0' => false,
    'phpdoc_summary' => false,
    'phpdoc_annotation_without_dot' => false,
    'phpdoc_order' => true,
    'array_syntax' => ['syntax' => 'short'],
    'ordered_imports' => true,
    'simplified_null_return' => false,
    'header_comment' => ['header' => $header],
    'yoda_style' => null,
    'native_function_invocation' => ['include' => ['@compiler_optimized']],
    'no_superfluous_phpdoc_tags' => true,
];

```

## Embed our templates in a PHP project

### Create a PHP project

For the sake of our example, we are going to create a blank PHP project, but you can of course skip this step if you already have a current PHP project that uses `PHP-CS-fixer`.

```shell
mkdir ~/my-php-project
cd ~/my-php-project
mkdir ./src
# Let's create a PHP file to five some food to PHP-CS-fixer
echo "<?php\n echo \"Coucou\";\n" > ./src/hello.php
composer init
composer require friendsofphp/php-cs-fixer
touch ./.php_cs.dist
```

Add the following content in `./php_cs.dist`:

```php
<?php

$header = <<<'EOF'
This file is part of the My-wonderful-project project.

Copyright © My company

@author My company <contact@my-company.com>
EOF;

$finder = PhpCsFixer\Finder::create()
    ->in([
        // App
        __DIR__ . '/src',
    ])
;

return PhpCsFixer\Config::create()
    ->setUsingCache(true)
    ->setRiskyAllowed(true)
    ->setFinder($finder)
    ->setRules([
        '@Symfony' => true,
        'psr0' => false,
        'phpdoc_summary' => false,
        'phpdoc_annotation_without_dot' => false,
        'phpdoc_order' => true,
        'array_syntax' => ['syntax' => 'short'],
        'ordered_imports' => true,
        'simplified_null_return' => false,
        'header_comment' => ['header' => $header],
        'yoda_style' => null,
        'native_function_invocation' => ['include' => ['@compiler_optimized']],
        'no_superfluous_phpdoc_tags' => true,
    ])
;
```

:bulb: We have hard-coded our coding rules but in the next step, we will of course replace them with our shared rules.

Run `vendor/bin/php-cs-fixer fix --dry-run` to check that your PHP-CS-fix config is OK.

### Embed our PHP templates in our PHP project

Create a `.manala.yaml` at the root of your PHP project:

`touch ./.manala.yaml`

And add the following content:

```yaml
manala:
  repository: /path/to/your/home/my-manala-template-repository
  template: my-php-templates
```

> :warning: Update `/path/to/your/home/` to match your real home !!! Using `~` won't work !!!

And finally run the following command:


```shell
manala up
# More verbose:
# manala up --debug
```

This command should have created a `.manala` folder at the root of your project, including a `php-cs-rules.php` file.

