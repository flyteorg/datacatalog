from setuptools import setup, find_packages

__version__ = '0.0.1b3'

setup(
    name='flyte-datacatalog',
    version=__version__,
    description='Proto IDL for Data Catalog Service',
    url='https://www.github.com/lyft/datacatalog',
    maintainer='Andrew Chan',
    maintainer_email='flyte-eng@lyft.com',
    packages=find_packages('protos/gen/pb_python'),
    package_dir={'': 'protos/gen/pb_python'},
    dependency_links=[],
    install_requires=[
        'protobuf>=3.5.0,<4.0.0',
        # Packages in here should rarely be pinned. This is because these
        # packages (at the specified version) are required for project
        # consuming this library. By pinning to a specific version you are the
        # number of projects that can consume this or forcing them to
        # upgrade/downgrade any dependencies pinned here in their project.
        #
        # Generally packages listed here are pinned to a major version range.
        #
        # e.g.
        # Python FooBar package for foobaring
        # pyfoobar>=1.0, <2.0
        #
        # This will allow for any consuming projects to use this library as
        # long as they have a version of pyfoobar equal to or greater than 1.x
        # and less than 2.x installed.
        "flyteidl>=0.14.0,<1.0.0",
    ],
    extras_require={
        ':python_version=="2.7"': ['typing>=3.6'],  # allow typehinting PY2
    },
    license="apache2",
    python_requires=">=2.7"
)
